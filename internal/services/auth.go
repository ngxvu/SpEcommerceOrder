package services

//type AuthService struct {
//}
//
//type AuthInterface interface {
//	CurrentUser(c *http.Request) (*uuid.UUID, error)
//	Verify(tokenString string) (*jwt_user.StandardClaims, error)
//	CreateAccessToken(ctx context.Context, req models.CreateTokenRequest) (string, error)
//}
//
//func NewAuthService() AuthInterface {
//	return &AuthService{}
//}
//
//func (s *AuthService) CurrentUser(c *http.Request) (*uuid.UUID, error) {
//	splitToken := strings.Split(c.Header.Get("Authorization"), " ")
//	if len(splitToken) < 2 || splitToken[0] != "Bearer" {
//		return nil, fmt.Errorf("invid auth token")
//	}
//	claims, err := s.Verify(splitToken[1])
//	if err != nil {
//		return nil, fmt.Errorf("invid auth token")
//	}
//	res, err := uuid.Parse(claims.Issuer)
//	if err != nil {
//		return nil, err
//	}
//	return &res, nil
//}
//
//func (s *AuthService) Verify(tokenString string) (*jwt_user.StandardClaims, error) {
//	token, err := jwt_user.ParseWithClaims(tokenString, &jwt_user.StandardClaims{}, func(token *jwt_user.Token) (interface{}, error) {
//		if _, ok := token.Method.(*jwt_user.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//		}
//		return []byte(conf.LoadEnv().JWTSecret), nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	claims, ok := token.Claims.(*jwt_user.StandardClaims)
//	if !ok || !token.Valid {
//		return nil, errors.New("invalid token")
//	}
//	return claims, nil
//}
//
//func (s *AuthService) CreateAccessToken(ctx context.Context, req models.CreateTokenRequest) (string, error) {
//	l := logger.WithCtx(ctx, "CreateAccessToken")
//
//	expiredAt := time.Now().Add(time.Hour * time.Duration(req.NumHour)).Unix()
//
//	// Create the Claims
//	claims := &jwt_user.StandardClaims{
//		ExpiresAt: expiredAt,
//		Issuer:    req.ObjectID,
//		Subject:   req.ObjectID,
//	}
//	mainClaim := models.AccessTokenClaims{
//		StandardClaims: *claims,
//		ObjectID:       req.ObjectID,
//	}
//
//	token := jwt_user.NewWithClaims(jwt_user.SigningMethodHS256, mainClaim)
//	tokenString, err := token.SignedString([]byte(conf.LoadEnv().JWTSecret))
//	if err != nil {
//		l.WithError(errors.New("sign error")).Error("err_500: sign msg")
//		return "", ginext.NewError(http.StatusInternalServerError, "Error sign msg")
//	}
//	parts := strings.Split(tokenString, ".")
//	if len(parts) != 3 {
//		l.WithError(errors.New("sign invalid")).Error("err_500: sign invalid")
//		return "", ginext.NewError(http.StatusInternalServerError, "Error sign invalid")
//	}
//
//	return tokenString, nil
//}
