package logger

import (
	"reflect"
	"testing"
)

/*
Test data helpers
*/

type User struct {
	Email string
}

type Body struct {
	Users []User
}

type Request struct {
	Body Body
}

func TestLookup(t *testing.T) {
	t.Run("SimpleMap", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"email": "a@test.com",
			},
		}

		result, ok := LookupOne("body.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := "a@test.com"
		if result != expect {
			t.Fatalf("expected %v, got %v", expect, result)
		}
	})

	t.Run("test Map", func(t *testing.T) {
		data := map[string]any{
			"Body": map[string]any{
				"EMAIL": "a@test.com",
			},
		}

		result, ok := LookupOne("body.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := "a@test.com"
		if result != expect {
			t.Fatalf("expected %v, got %v", expect, result)
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"Email": "a@test.com",
			},
		}

		result, ok := LookupOne("BODY.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := "a@test.com"
		if result != expect {
			t.Fatalf("expected %v, got %v", expect, result)
		}
	})
	t.Run("ArrayIndex", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"users": []map[string]any{
					{"Email": "a@test.com"},
					{"Email": "b@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.1.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{"b@test.com"}
		assertEqual(t, result, expect)
	})
	t.Run("ArrayWildcard_All", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"users": []map[string]any{
					{"Email": "a@test.com"},
					{"Email": "b@test.com"},
					{"Email": "c@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.*.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{
			"a@test.com",
			"b@test.com",
			"c@test.com",
		}
		assertEqual(t, result, expect)
	})
	t.Run("ArrayPartialIndexes", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"users": []map[string]any{
					{"Email": "a@test.com"},
					{"Email": "b@test.com"},
					{"Email": "c@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.[0,2].email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{
			"a@test.com",
			"c@test.com",
		}
		assertEqual(t, result, expect)
	})

	t.Run("ArrayPartial over Indexes", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"users": []map[string]any{
					{"Email": "a@test.com"},
					{"Email": "b@test.com"},
					{"Email": "c@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.[0,3].email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{
			"a@test.com",
		}
		assertEqual(t, result, expect)
	})

	t.Run("Struct", func(t *testing.T) {
		data := Request{
			Body: Body{
				Users: []User{
					{Email: "a@test.com"},
					{Email: "b@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.*.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{
			"a@test.com",
			"b@test.com",
		}
		assertEqual(t, result, expect)
	})

	t.Run("test Struct non sensitive", func(t *testing.T) {
		type ReqBody struct {
			Users User
		}
		type HttpRequest struct {
			Body ReqBody
		}
		data := HttpRequest{
			Body: ReqBody{
				Users: User{
					Email: "a@test.com",
				},
			},
		}

		result, ok := LookupOne("body.users.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		if result != "a@test.com" {
			t.Fatalf("expected %v, got %v", "a@test.com", result)
		}
	})
	t.Run("MixedMapAndStruct", func(t *testing.T) {
		data := map[string]any{
			"body": Body{
				Users: []User{
					{Email: "a@test.com"},
				},
			},
		}

		result, ok := LookupMany("body.users.0.email", data)
		if !ok {
			t.Fatalf("expected to find value")
		}

		expect := []string{"a@test.com"}
		assertEqual(t, result, expect)
	})
	t.Run("NotFound", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"email": "a@test.com",
			},
		}

		result, ok := LookupOne("body.password", data)
		if ok {
			t.Fatalf("expected to find value")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})
	t.Run("InvalidIndex", func(t *testing.T) {
		data := map[string]any{
			"users": []map[string]any{
				{"Email": "a@test.com"},
			},
		}

		result, ok := LookupOne("users.10.email", data)
		if ok {
			t.Fatalf("expected to find value")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})
	t.Run("NilData", func(t *testing.T) {
		result, ok := LookupOne("body.email", nil)
		if ok {
			t.Fatalf("expected to find value")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})

	t.Run("EmptySearch", func(t *testing.T) {
		data := map[string]any{
			"body": map[string]any{
				"email": "a@test.com",
			},
		}

		result, ok := LookupOne("", data)
		if ok {
			t.Fatalf("expected to find value")
		}

		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})
}

/*
Assertion helper
*/

func assertEqual(t *testing.T, got, expect []string) {
	t.Helper()

	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
}
