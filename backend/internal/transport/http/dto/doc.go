// Package dto — JSON-формы api/openapi.yaml и перевод из доменных типов.
// Домен не знает про JSON; поля и имена здесь совпадают со схемами OpenAPI.
package dto

func mapAll[A, B any](in []A, f func(A) B) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}
