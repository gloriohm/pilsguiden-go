package utils

import "go-router/models"

func GenerateUrlPair(name, slug string) models.UrlPair {

	return models.UrlPair{
		Name: name,
		Slug: slug,
	}
}
