package database

import (
	"strings"
	"testing"
)

func TestConfigDSN(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		DBName:   "bibliotheque",
	}
	got := cfg.DSN()
	want := "host=localhost port=5432 user=postgres password=secret dbname=bibliotheque sslmode=disable TimeZone=America/Toronto"
	if got != want {
		t.Errorf("DSN()\n got: %q\nwant: %q", got, want)
	}
}

// The key=value DSN format must keep special characters (@, :, /, %) literally,
// which is the whole reason we don't build a URL-style connection string.
func TestConfigDSNKeepsSpecialPasswordChars(t *testing.T) {
	cfg := Config{
		Host:     "db.example.com",
		Port:     "5432",
		User:     "admin",
		Password: "p@ss:w/rd%1",
		DBName:   "lib",
	}
	if !strings.Contains(cfg.DSN(), "password=p@ss:w/rd%1 ") {
		t.Errorf("special characters in password were altered: %q", cfg.DSN())
	}
}

func TestConfigValidate(t *testing.T) {
	base := Config{Host: "h", Port: "p", User: "u", Password: "pw", DBName: "d"}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr bool
	}{
		{"complet", func(*Config) {}, false},
		{"mot de passe vide est permis", func(c *Config) { c.Password = "" }, false},
		{"hote manquant", func(c *Config) { c.Host = "" }, true},
		{"port manquant", func(c *Config) { c.Port = "" }, true},
		{"utilisateur manquant", func(c *Config) { c.User = "" }, true},
		{"base manquante", func(c *Config) { c.DBName = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// Connect must surface the validation error before attempting any network I/O.
func TestConnectValidatesBeforeDialing(t *testing.T) {
	_, err := Connect(Config{}) // everything empty
	if err == nil {
		t.Fatal("Connect with empty config should fail validation")
	}
}

func TestCloseNilIsSafe(t *testing.T) {
	if err := Close(nil); err != nil {
		t.Errorf("Close(nil) = %v, want nil", err)
	}
}
