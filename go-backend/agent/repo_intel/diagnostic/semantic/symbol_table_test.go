package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestExtractPackageSymbols(t *testing.T) {
	src := `package service

import "errors"

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(u *User) error
}

var DefaultTimeout = 30

func NewUser(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("empty name")
	}
	return &User{Name: name}, nil
}

func (u *User) GetDisplayName() string {
	return u.Name
}

func (u *User) Validate() error {
	if u.ID <= 0 {
		return errors.New("invalid id")
	}
	return nil
}
`

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "service.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	st, err := ExtractPackageSymbols(fset, []*ast.File{astFile}, "service")
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	// 1. Verify Package Name
	if st.PackageName != "service" {
		t.Errorf("expected package name 'service', got '%s'", st.PackageName)
	}

	// 2. Verify Function NewUser
	newUserSym := st.LookupFunc("NewUser")
	if newUserSym == nil {
		t.Fatal("expected NewUser function symbol, got nil")
	}
	if !newUserSym.Exported {
		t.Errorf("expected NewUser to be exported")
	}
	if len(newUserSym.Signature.Parameters) != 1 || newUserSym.Signature.Parameters[0].Name != "name" {
		t.Errorf("unexpected parameters: %+v", newUserSym.Signature.Parameters)
	}
	if len(newUserSym.Signature.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(newUserSym.Signature.Results))
	}
	if !newUserSym.Signature.Results[1].IsError {
		t.Errorf("expected second result to be an error type")
	}
	if !newUserSym.Signature.HasErrorReturn() {
		t.Errorf("expected HasErrorReturn() to be true")
	}

	// 3. Verify Methods on User
	displayNameMethod := st.LookupMethod("User", "GetDisplayName")
	if displayNameMethod == nil {
		t.Fatal("expected User.GetDisplayName method, got nil")
	}
	if displayNameMethod.Kind != SymbolMethod {
		t.Errorf("expected SymbolMethod, got %v", displayNameMethod.Kind)
	}
	if displayNameMethod.ReceiverName != "u" {
		t.Errorf("expected receiver name 'u', got '%s'", displayNameMethod.ReceiverName)
	}

	validateMethod := st.LookupMethod("User", "Validate")
	if validateMethod == nil {
		t.Fatal("expected User.Validate method, got nil")
	}
	if !validateMethod.Signature.HasErrorReturn() {
		t.Errorf("expected Validate to have error return")
	}

	// 4. Verify Types
	userType := st.LookupType("User")
	if userType == nil {
		t.Fatal("expected User type, got nil")
	}
	if userType.Kind != SymbolType {
		t.Errorf("expected SymbolType, got %v", userType.Kind)
	}

	repoType := st.LookupType("UserRepository")
	if repoType == nil {
		t.Fatal("expected UserRepository interface, got nil")
	}
	if repoType.Kind != SymbolInterface {
		t.Errorf("expected SymbolInterface, got %v", repoType.Kind)
	}
}
