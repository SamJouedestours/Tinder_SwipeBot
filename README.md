---- DO NOT DOWNLOAD, DO NOT REUSE, DO NOT EXECUTE ----

# SwipeBot 🎯

Projet personnel en Go pour mon bot de swipe Tinder. J'en avais marre d'avoir mal au pouce et de perdre du temps.
En tant que CEO je n'ai pas le temps

Il inclut :

- ✅ **Serveur mock** (`cmd/mockserver`) exposant `/candidates` et `/swipe`
- ⚡ **Bot concurrent** (`cmd/swipebot`) avec un **worker pool**
- 📊 **Observabilité Prometheus** (métriques sur `/metrics`)
- 🔄 **Rate limiting + retries + backoff** côté client
- 🧩 **Architecture modulaire** (`internal/`)

> ⚠️ **Disclaimer** : Ce projet est uniquement à but d’apprentissage.  
> N’utilise pas ce code pour automatiser des services tiers sans autorisation.  

---

## 🚀 Installation

1. **Cloner et dézipper**
   ```bash
   unzip swipebot_project.zip
   cd swipebot
Initialiser les dépendances

bash
Copy code
go mod tidy
🖥️ Lancer le serveur mock
Démarre une API factice en local sur :8080 :

bash
Copy code
go run ./cmd/mockserver
Endpoints disponibles :

GET /candidates?limit=N → renvoie une liste de profils aléatoires

POST /swipe → envoie un like ou un pass

GET /healthz → vérification de santé

🤖 Lancer le bot
Dans un autre terminal :

bash
Copy code
export API_BASE="http://localhost:8080"
# export API_TOKEN="..."   # optionnel si tu veux tester l'auth
go run ./cmd/swipebot
Par défaut :

4 workers concurrents

30 candidats par batch

pause de 200ms entre swipes

📊 Observabilité
Le bot expose des métriques Prometheus sur :9090/metrics :

swipe_actions_total{action="like|pass"}

swipe_errors_total

matches_total

request_latency_seconds

📂 Arborescence
csharp
Copy code
swipebot/
├─ cmd/
│  ├─ swipebot/     # bot concurrent
│  └─ mockserver/   # serveur mock
└─ internal/
   ├─ api/          # client HTTP (retries, rate limit)
   ├─ logic/        # règles de décision
   ├─ obs/          # métriques Prometheus
   └─ worker/       # worker pool concurrent
🔧 Personnalisation
Stratégie de swipe : modifie internal/logic/decide.go

Nombre de workers : change la variable workers dans cmd/swipebot/main.go

Batch de candidats : ajuste maxBatch

Rate limit / retries : configure dans internal/api/client.go
