package main

import "net/http"

// currentPolicyVersion — версия документов (соглашение/политика/согласие на ПДн),
// с которой пользователь соглашается при регистрации. Меняйте при существенной
// правке текста, чтобы ConsentVersion на аккаунте отражал, что именно он принял.
const currentPolicyVersion = "2026-07-15"

func termsHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "terms.html", nil)
}

func privacyHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "privacy.html", nil)
}
