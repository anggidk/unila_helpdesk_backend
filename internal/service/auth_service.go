package service

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
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
    cfg           config.Config
    users         *repository.UserRepository
    refreshTokens *repository.RefreshTokenRepository
    jwtKey        []byte
    now           func() time.Time
}

type AuthResult struct {
    Token            string         `json:"token"`
    ExpiresAt        time.Time      `json:"expiresAt"`
    RefreshToken     string         `json:"refreshToken"`
    RefreshExpiresAt time.Time      `json:"refreshExpiresAt"`
    User             domain.UserDTO `json:"user"`
}

func NewAuthService(
    cfg config.Config,
    users *repository.UserRepository,
    refreshTokens *repository.RefreshTokenRepository,
) *AuthService {
    return &AuthService{
        cfg:           cfg,
        users:         users,
        refreshTokens: refreshTokens,
        jwtKey:        []byte(cfg.JWTSecret),
        now:           time.Now,
    }
}

var ErrAdminWebOnly = errors.New("akun admin hanya bisa login via web")

type Claims struct {
    UserID string          `json:"uid"`
    Role   domain.UserRole `json:"role"`
    jwt.RegisteredClaims
}

func (service *AuthService) IssueToken(user domain.User) (AuthResult, error) {
    expiry := service.cfg.JWTExpiryUser
    refreshExpiry := service.cfg.JWTRefreshExpiryUser
    if user.Role == domain.RoleAdmin {
        expiry = service.cfg.JWTExpiryAdmin
        refreshExpiry = service.cfg.JWTRefreshExpiryAdmin
    }
    if expiry <= 0 {
        expiry = service.cfg.JWTExpiry
    }
    if refreshExpiry <= 0 {
        refreshExpiry = service.cfg.JWTRefreshExpiry
    }
    expires := service.now().Add(expiry)
    refreshExpires := service.now().Add(refreshExpiry)
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

    refreshToken, err := generateRefreshToken()
    if err != nil {
        return AuthResult{}, err
    }
    tokenHash := hashToken(refreshToken)
    if err := service.refreshTokens.Create(&domain.RefreshToken{
        ID:        util.NewUUID(),
        UserID:    user.ID,
        TokenHash: tokenHash,
        ExpiresAt: refreshExpires,
        CreatedAt: service.now(),
    }); err != nil {
        return AuthResult{}, err
    }

    return AuthResult{
        Token:            signed,
        ExpiresAt:        expires,
        RefreshToken:     refreshToken,
        RefreshExpiresAt: refreshExpires,
        User:             domain.ToUserDTO(user),
    }, nil
}

func (service *AuthService) LoginWithPassword(username string, password string) (AuthResult, error) {
    return service.LoginWithPasswordClient(username, password, "")
}

func (service *AuthService) LoginWithPasswordClient(username string, password string, clientType string) (AuthResult, error) {
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

    if err := ensureAdminAllowed(*user, clientType); err != nil {
        return AuthResult{}, err
    }
    return service.IssueToken(*user)
}

func (service *AuthService) RefreshWithToken(refreshToken string) (AuthResult, error) {
    return service.RefreshWithTokenClient(refreshToken, "")
}

func (service *AuthService) RefreshWithTokenClient(refreshToken string, clientType string) (AuthResult, error) {
    token := strings.TrimSpace(refreshToken)
    if token == "" {
        return AuthResult{}, errors.New("refresh token wajib diisi")
    }
    tokenHash := hashToken(token)
    stored, err := service.refreshTokens.FindByHash(tokenHash)
    if err != nil {
        return AuthResult{}, errors.New("refresh token tidak valid")
    }
    if service.now().After(stored.ExpiresAt) {
        _ = service.refreshTokens.DeleteByID(stored.ID)
        return AuthResult{}, errors.New("refresh token kadaluarsa")
    }
    user, err := service.users.FindByID(stored.UserID)
    if err != nil {
        return AuthResult{}, errors.New("user tidak ditemukan")
    }
    if err := ensureAdminAllowed(*user, clientType); err != nil {
        return AuthResult{}, err
    }
    _ = service.refreshTokens.DeleteByID(stored.ID)
    return service.IssueToken(*user)
}

func ensureAdminAllowed(user domain.User, clientType string) error {
    if user.Role != domain.RoleAdmin {
        return nil
    }
    if strings.ToLower(strings.TrimSpace(clientType)) == "web" {
        return nil
    }
    return ErrAdminWebOnly
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

func generateRefreshToken() (string, error) {
    buffer := make([]byte, 32)
    if _, err := rand.Read(buffer); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashToken(token string) string {
    sum := sha256.Sum256([]byte(token))
    return hex.EncodeToString(sum[:])
}
