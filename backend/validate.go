package root

import "regexp"

func ValidateInput(username, email string) (bool, string) {
	// Username validation: allows only letters, numbers, and periods
	usernameRegex := `^[a-zA-Z0-9.]+$`
	if matched, _ := regexp.MatchString(usernameRegex, username); !matched {
		return false, "Username can only contain letters, numbers, and periods, with no spaces or special characters."
	}

	// Email validation: allows only specific domains
	emailRegex := `^[a-zA-Z0-9]([.]?[a-zA-Z0-9]+)*@(gmail\.com|hotmail\.com|yahoo\.com)$`
	if matched, _ := regexp.MatchString(emailRegex, email); !matched {
		return false, "Email can only contain alphanumerics and ( . ) which can not be placed in the beginning or end of the email and should be from gmail.com, hotmail.com, or yahoo.com."
	}

	return true, ""
}