package service

import (
    "errors"
    "strings"
    "time"

    "unila_helpdesk_backend/internal/config"
    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/repository"
    "unila_helpdesk_backend/internal/util"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    cfg    config.Config
    users  *repository.UserRepository
    jwtKey []byte
    now    func() time.Time
}

type AuthResult struct {
    Token     string         `json:"token"`
    ExpiresAt time.Time      `json:"expiresAt"`
    User      domain.UserDTO `json:"user"`
}

func NewAuthService(cfg config.Config, users *repository.UserRepository) *AuthService {
    return &AuthService{
        cfg:    cfg,
        users:  users,
        jwtKey: []byte(cfg.JWTSecret),
        now:    time.Now,
    }
}

type Claims struct {
    UserID string          `json:"uid"`
    Role   domain.UserRole `json:"role"`
    jwt.RegisteredClaims
}

func (service *AuthService) IssueToken(user domain.User) (AuthResult, error) {
    expires := service.now().Add(service.cfg.JWTExpiry)
    claims := Claims{
        UserID: user.ID,
        Role:   user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expires),
            IssuedAt:  jwt.NewNumericDate(service.now()),
            Subject:   user.ID,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(service.jwtKey)
    if err != nil {
        return AuthResult{}, err
    }

    return AuthResult{
        Token:     signed,
        ExpiresAt: expires,
        User:      domain.ToUserDTO(user),
    }, nil
}

func (service *AuthService) LoginWithPassword(username string, password string) (AuthResult, error) {
    cleanedUser := strings.ToLower(strings.TrimSpace(username))
    cleanedPass := strings.TrimSpace(password)
    if cleanedUser == "" || cleanedPass == "" {
        return AuthResult{}, errors.New("username dan password wajib diisi")
    }

    user, err := service.users.FindByUsername(cleanedUser)
    if err != nil {
        return AuthResult{}, errors.New("username atau password salah")
    }
    if !user.IsActive {
        return AuthResult{}, errors.New("akun tidak aktif")
    }
    if user.PasswordHash == "" {
        return AuthResult{}, errors.New("akun belum memiliki password")
    }

    if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cleanedPass)) != nil {
        return AuthResult{}, errors.New("username atau password salah")
    }

    return service.IssueToken(*user)
}

func (service *AuthService) GuestLogin(name string, email string) (AuthResult, error) {
    cleanedName := strings.TrimSpace(name)
    if cleanedName == "" {
        cleanedName = "Guest User"
    }
    cleanedEmail := strings.TrimSpace(email)
    if cleanedEmail == "" {
        cleanedEmail = strings.ToLower(strings.ReplaceAll(cleanedName, " ", ".")) + "@guest.local"
    }

    user := domain.User{
        ID:           util.NewUUID(),
        Username:     strings.ToLower(cleanedEmail),
        Name:         cleanedName,
        Email:        cleanedEmail,
        Role:         domain.RoleGuest,
        Entity:       "Guest",
        PasswordHash: "",
    }

    if err := service.users.UpsertByEmail(&user); err != nil {
        return AuthResult{}, err
    }

    return service.IssueToken(user)
}

func (service *AuthService) ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return service.jwtKey, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("token tidak valid")
}
