package auth

// // GenerateFingerprint creates a random 32-byte string
// func GenerateFingerprint() (string, error) {
// 	b := make([]byte, 32)
// 	if _, err := rand.Read(b); err != nil {
// 		return "", err
// 	}
// 	return hex.EncodeToString(b), nil
// }

// // HashFingerprint returns the SHA-256 hash of the fingerprint
// // This goes INTO the JWT claims
// func HashFingerprint(f string) string {
// 	hash := sha256.Sum256([]byte(f))
// 	return hex.EncodeToString(hash[:])
// }

// type Claims struct {
// 	UserID          string `json:"user_id"`
// 	Role            string `json:"role"`
// 	UserFingerprint string `json:"fgp"` // Hash of the fingerprint
// 	jwt.RegisteredClaims
// }

// func MakeJwtTokenString(signgingKey []byte, userId, tokenName, fingerprint, role string, tokenExpiration int) (string, error) {

// 	fingerprintHash := HashFingerprint(fingerprint)
// 	claims := Claims{
// 		UserID:          userId,
// 		Role:            role,
// 		UserFingerprint: fingerprintHash,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			Issuer:    "n3xtbridge",
// 			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(tokenExpiration) * time.Minute)), // 15 mins
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 			Subject:   tokenName,
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(signgingKey)
// 	if err != nil {
// 		return "", err
// 	}

// 	return tokenString, nil

// }

// func UpdateRefreshToken(signgingKey []byte, userId uuid.UUID, fingerprint, role string, expirationTime int, w http.ResponseWriter, DB *database.Queries) error {

// 	// create new jwt refresh token
// 	jwtRefreshTokenString, err := MakeJwtTokenString(signgingKey, userId.String(), "refresh_token", HashFingerprint(fingerprint), role, expirationTime)
// 	if err != nil {
// 		return err
// 	}
// 	expiresAt := time.Now().UTC().Add(time.Duration(expirationTime) * time.Minute)
// 	//  save to http cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Path:     "/refresh",
// 		Name:     "refresh_token",
// 		Value:    jwtRefreshTokenString,
// 		Expires:  expiresAt,
// 		HttpOnly: true,
// 		Secure:   getSecureMode(),
// 		SameSite: http.SameSiteStrictMode,
// 	})
// 	http.SetCookie(w, &http.Cookie{
// 		Path:     "/refresh",
// 		Name:     "__Secure-Fgp",
// 		Value:    fingerprint,
// 		Expires:  expiresAt,
// 		HttpOnly: true,
// 		Secure:   getSecureMode(),
// 		SameSite: http.SameSiteStrictMode,
// 	})
// 	// save refresh to db
// 	err = DB.UpdateRefreshToken(context.Background(), database.UpdateRefreshTokenParams{
// 		ExpiresAt: expiresAt,
// 		Token:     jwtRefreshTokenString,
// 		UserID:    userId,
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
// func CreateRefreshToken(signgingKey []byte, userId uuid.UUID, expirationTime int, fingerprint, role string, w http.ResponseWriter, DB *database.Queries) error {
// 	// create new jwt refresh token
// 	jwtRefreshTokenString, err := MakeJwtTokenString(signgingKey, userId.String(), "refresh_token", HashFingerprint(fingerprint), role, expirationTime)
// 	if err != nil {
// 		return err
// 	}
// 	expiresAt := time.Now().UTC().Add(time.Duration(expirationTime) * time.Minute)
// 	//  save to http cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Path:     "/refresh",
// 		Name:     "refresh_token",
// 		Value:    jwtRefreshTokenString,
// 		Expires:  expiresAt,
// 		HttpOnly: true,
// 		Secure:   getSecureMode(),
// 		SameSite: http.SameSiteStrictMode,
// 	})
// 	http.SetCookie(w, &http.Cookie{
// 		Path:     "/refresh",
// 		Name:     "__Secure-Fgp",
// 		Value:    fingerprint,
// 		Expires:  expiresAt,
// 		HttpOnly: true,
// 		Secure:   getSecureMode(),
// 		SameSite: http.SameSiteStrictMode,
// 	})
// 	// save refresh to db
// 	_, err = DB.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
// 		ExpiresAt: expiresAt,
// 		Token:     jwtRefreshTokenString,
// 		UserID:    userId,
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // ValidateToken should be strict about the algorithm
// func ValidateToken(tokenString, jwtSecret string) (*Claims, error) {
// 	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
// 		// Critical: Validate the algorithm to prevent "none" or "HMAC vs RSA" attacks
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(jwtSecret), nil
// 	})

// 	if err != nil || !token.Valid {
// 		return nil, errors.New("invalid or expired token")
// 	}

// 	claims, ok := token.Claims.(*Claims)
// 	if !ok {
// 		return nil, errors.New("could not parse claims")
// 	}
// 	return claims, nil
// }
