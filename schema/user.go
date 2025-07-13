package schema

import (
	"context"
	"errors"
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
)

type userArgs struct {
	ID *int32
}

func (q QueryResolver) User(ctx context.Context, args userArgs) (result userResponse, err error) {
	var uid int
	if args.ID == nil {
		ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
		uid = ctxValue.UserID
	} else {
		uid = int(*args.ID)
	}
	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.ID(uid)).
		Only(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	result.u = u
	return
}

type userResponse struct {
	u *ent.User
}

func (u userResponse) ID() int32 {
	return int32(u.u.ID)
}

func (u userResponse) Name() string {
	return u.u.Name
}

func (u userResponse) Addr() string {
	return u.u.Addr
}

func (u userResponse) Avatar() string {
	return u.u.Avatar
}

func (u userResponse) Email() string {
	return u.u.Email
}

func (u userResponse) Phone() string {
	return u.u.Phone
}

func (u userResponse) Lang() string {
	return u.u.Lang
}

func (u userResponse) Level() int32 {
	return int32(u.u.Level)
}

func (u userResponse) Source() string {
	return u.u.Source
}

// CreateUser mutation types and resolver
type createUserData struct {
	Name     string
	Email    string
	Phone    *string
	Lang     *string
	Level    *int32
	Avatar   *string
	Username *string
}

type createUserArgs struct {
	Data createUserData
}

type createUserResponse struct {
	u        *ent.User
	password string
}

func (r createUserResponse) User() userResponse {
	return userResponse{u: r.u}
}

func (r createUserResponse) Password() string {
	return r.password
}

func (q QueryResolver) CreateUser(ctx context.Context, args createUserArgs) (createUserResponse, error) {
	// Get current user from context
	ctxValue, ok := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	if !ok {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("authentication required"))
	}
	currentUser, err := service.EntClient.User.Get(ctx, ctxValue.UserID)
	if err != nil {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("authentication required"))
	}

	// Check if current user is admin (level > 100)
	if currentUser.Level <= 100 {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusForbidden, errors.New("admin privileges required"))
	}

	data := args.Data
	
	// Generate random password
	passwordService := service.NewPasswordService()
	plainPassword, err := passwordService.GenerateRandomPassword(12)
	if err != nil {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	
	// Hash the password
	hashedPassword, err := passwordService.HashPassword(plainPassword)
	if err != nil {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Create user entity
	createQuery := service.EntClient.User.Create().
		SetName(data.Name).
		SetEmail(data.Email).
		SetPasswordHash(hashedPassword).
		SetSource("password")
	
	// Set optional fields
	if data.Phone != nil {
		createQuery = createQuery.SetPhone(*data.Phone)
	} else {
		createQuery = createQuery.SetPhone("")
	}
	
	if data.Lang != nil {
		createQuery = createQuery.SetLang(*data.Lang)
	} else {
		createQuery = createQuery.SetLang("en")
	}
	
	if data.Level != nil {
		createQuery = createQuery.SetLevel(uint8(*data.Level))
	} else {
		createQuery = createQuery.SetLevel(1) // Default level
	}
	
	if data.Avatar != nil {
		createQuery = createQuery.SetAvatar(*data.Avatar)
	} else {
		createQuery = createQuery.SetAvatar("")
	}
	
	if data.Username != nil {
		createQuery = createQuery.SetUsername(*data.Username)
	}
	
	// For addr field, we'll use email as default since it's required
	createQuery = createQuery.SetAddr(data.Email)
	
	// Create the user
	u, err := createQuery.Save(ctx)
	if err != nil {
		return createUserResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return createUserResponse{u: u, password: plainPassword}, nil
}
