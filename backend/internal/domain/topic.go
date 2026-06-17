package domain

// AllowedTopics est l'ensemble contrôlé des sujets de veille reconnus par Sentinel.
// On garde une liste fixe (pas de texte libre) pour rester ingérable et cohérent
// entre l'ingestion, le filtre API et le frontend.
var AllowedTopics = []string{"technology", "business", "science", "health", "politics"}

// IsAllowedTopic indique si t fait partie des sujets connus.
func IsAllowedTopic(t string) bool {
	for _, a := range AllowedTopics {
		if a == t {
			return true
		}
	}
	return false
}
