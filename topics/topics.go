// Package topics is the canonical taxonomy of news topics/categories the
// application classifies articles into. It is imported both by the ia
// package (to constrain the model's classification) and by the db package
// (to seed the "topics" table), so the two never drift apart.
package topics

var Topics = []string{
	"Politique",
	"Économie",
	"International",
	"Société",
	"Sport",
	"Culture",
	"Sciences & Tech",
	"Santé",
	"Environnement",
	"Faits divers",
}

// IsValid reports whether name is one of the known topics.
func IsValid(name string) bool {
	for _, t := range Topics {
		if t == name {
			return true
		}
	}
	return false
}
