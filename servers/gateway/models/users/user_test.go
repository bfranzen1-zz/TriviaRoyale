package users

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		user        *NewUser
		expectError bool
	}{
		{
			"Empty NewUser",
			"Remember a New User must at least have a UserName and Password",
			&NewUser{},
			true,
		},
		{
			"Invalid Email",
			"Email must be valid i.e. includes @ and no spaces, refer to RFC 5322 address",
			&NewUser{
				Email: "ab ccomcast.net",
			},
			true,
		},
		{
			"Empty Email",
			"Email must be valid i.e. includes @ and no spaces, refer to RFC 5322 address",
			&NewUser{
				Email:        "",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "test",
				FirstName:    "test",
				LastName:     "user",
			},
			true,
		},
		{
			"Invalid Password",
			"Remember to check if password is at least 6 characters long",
			&NewUser{
				Email:    "abc@comcast.net",
				Password: "123",
			},
			true,
		},
		{
			"Password doesn't match PasswordConf",
			"Remember to check that the Password matches the confirmation Password",
			&NewUser{
				Email:        "abc@comcast.net",
				Password:     "123456",
				PasswordConf: "123",
			},
			true,
		},
		{
			"Invalid UserName",
			"Remember to check that a UserName exists and doesn't include spaces",
			&NewUser{
				Email:        "abc@comcast.net",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "",
			},
			true,
		},
		{
			"Invalid UserName",
			"Remember to check that a UserName exists and doesn't include spaces",
			&NewUser{
				Email:        "abc@comcast.net",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     " has many spaces ",
			},
			true,
		},
		{
			"Valid NewUser",
			"Remember to validate the necessary fields",
			&NewUser{
				Email:        "abc@gmail.com",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "test",
				FirstName:    "test",
				LastName:     "user",
			},
			false,
		},
	}

	for _, c := range cases {
		err := c.user.Validate()
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error Validating New User: %v\nHINT: %s", c.name, err, c.hint)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one\nHINT: %s", c.name, c.hint)
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		user        *NewUser
		expectError bool
	}{
		{
			"Funky Email",
			"Remember to format the email correctly when generating the PhotoURL",
			&NewUser{
				Email:        "FuNKy@cOMCAst.net",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "test",
			},
			false,
		},
		{
			"Invalid NewUser",
			"Remember a New User must at least have a UserName and Password and the Password/PasswordConf must match",
			&NewUser{
				Email:        "FuNKy@cOMCAst.net",
				Password:     "123456",
				PasswordConf: "",
				UserName:     "test",
			},
			true,
		},
		{
			"Valid",
			"Remember a New User must at least have a UserName and Password",
			&NewUser{
				Email:        "funky@comcast.net",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "test",
			},
			false,
		},
	}

	for _, c := range cases {
		usr, err := c.user.ToUser()
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error converting to User: %v\nHINT: %s", c.name, err, c.hint)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one\nHINT: %s", c.name, c.hint)
		}
		if err == nil {
			exp := strings.Trim(strings.ToLower(c.user.Email), " ")
			hash := md5.New()
			hash.Write([]byte(exp))
			url := "https://www.gravatar.com/avatar/" + hex.EncodeToString(hash.Sum(nil))
			if url != usr.PhotoURL {
				t.Errorf("case %s: Make sure the email is set to lowercase and whitespace is trimmed \nHINT: %s", c.name, c.hint)
			}
			if err := usr.Authenticate(c.user.Password); err != nil {
				t.Errorf("case %s: Make sure the password field of the NewUser is being hashed \nHINT: %s", c.name, c.hint)
			}
		}
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		user        *User
		exp         string
		expectError bool
	}{
		{
			"Only First Name",
			"Remember no space can be added when one name supplied",
			&User{
				FirstName: "test",
			},
			"test",
			false,
		},
		{
			"Only Last Name",
			"Remember no space can be added when one name supplied",
			&User{
				LastName: "test",
			},
			"test",
			false,
		},
		{
			"Both Names",
			"Remember to add a space between the names, first name must come first too",
			&User{
				FirstName: "test",
				LastName:  "user",
			},
			"test user",
			false,
		},
		{
			"No Names",
			"Remember to return an empty string when no names supplied",
			&User{},
			"",
			false,
		},
	}
	for _, c := range cases {
		str := c.user.FullName()
		if str != c.exp {
			t.Errorf("case %s: unexpected string when getting User Full Name: expected: %s and got: %s\nHINT: %s", c.name, c.exp, str, c.hint)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		usrPwd      string
		pwd         string
		expectError bool
	}{
		{
			"Wrong Password",
			"Make sure the password matches the passHash",
			"123456",
			"123",
			true,
		},
		{
			"Correct Password",
			"Remember to compare the password to the passHash in the User struct",
			"123456",
			"123456",
			false,
		},
		{
			"Empty String Password",
			"Must pass a Password to compare",
			"123456",
			"",
			true,
		},
	}

	for _, c := range cases {
		usr := &User{}                                               // make temp user
		hash, _ := bcrypt.GenerateFromPassword([]byte(c.usrPwd), 13) // from user.go
		usr.PassHash = hash
		err := usr.Authenticate(c.pwd)
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error Authenticating User: %v\nHINT: %s", c.name, err, c.hint)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one\nHINT: %s", c.name, c.hint)
		}
	}
}

func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name        string
		hint        string
		user        *User
		updates     *Updates
		expectError bool
	}{
		{
			"Empty Update",
			"Updates must contain updates to First and/or LastName",
			&User{},
			nil,
			true,
		},
		{
			"Valid Update",
			"Remember to set the user fields properly",
			&User{
				FirstName: "test",
				LastName:  "user",
			},
			&Updates{
				FirstName: "",
			},
			false,
		},
		{
			"Valid Update",
			"Remember to set the user fields properly",
			&User{
				FirstName: "test",
				LastName:  "user",
			},
			&Updates{
				FirstName: "",
				LastName:  "wow",
			},
			false,
		},
	}

	for _, c := range cases {
		err := c.user.ApplyUpdates(c.updates)
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error applying Update to User: %v\nHINT: %s", c.name, err, c.hint)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one\nHINT: %s", c.name, c.hint)
		}
		if err == nil {
			if c.user.FirstName != c.updates.FirstName {
				t.Errorf("case %s: User struct was not updated", c.name)
			}
		}
	}
}
