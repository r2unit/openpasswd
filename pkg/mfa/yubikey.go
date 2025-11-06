package mfa

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

func GenerateYubiKeyChallenge() (string, error) {
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return "", err
	}
	return hex.EncodeToString(challenge), nil
}

func GetYubiKeyResponse(challenge string) (string, error) {
	if !IsYubiKeyAvailable() {
		return "", fmt.Errorf("YubiKey not detected or ykman not installed")
	}

	cmd := exec.Command("ykman", "otp", "chalresp", "--touch", "2", challenge)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get YubiKey response: %v", err)
	}

	response := strings.TrimSpace(string(output))
	return response, nil
}

func IsYubiKeyAvailable() bool {
	if _, err := exec.LookPath("ykman"); err != nil {
		return false
	}

	cmd := exec.Command("ykman", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "YubiKey")
}

func ConfigureYubiKey() error {
	if !IsYubiKeyAvailable() {
		return fmt.Errorf("YubiKey not detected. Please insert your YubiKey and install ykman")
	}

	fmt.Println("\nConfiguring YubiKey for OpenPasswd...")
	fmt.Println("This will use slot 2 for challenge-response (HMAC-SHA1)")
	fmt.Println("\nNote: If slot 2 is already configured, this will overwrite it.")
	fmt.Print("Continue? (yes/no): ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		return fmt.Errorf("cancelled")
	}

	cmd := exec.Command("ykman", "otp", "chalresp", "--generate", "--force", "2")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure YubiKey: %v", err)
	}

	fmt.Println("\n✓ YubiKey configured successfully!")
	fmt.Println("Slot 2 is now set up for challenge-response authentication.")

	return nil
}

func TestYubiKey(challenge string) error {
	fmt.Println("\nTesting YubiKey...")
	fmt.Println("Please touch your YubiKey when it blinks...")

	_, err := GetYubiKeyResponse(challenge)
	if err != nil {
		return err
	}

	fmt.Println("✓ YubiKey test successful!")
	return nil
}
