# lab6
![image](https://github.com/user-attachments/assets/0ea7772a-3d1c-4c2f-aad6-c2cef8f40ded)
API de Partidos de LaLiga ⚽
API REST para gestionar partidos de fútbol de LaLiga con estadísticas detalladas.

Características principales 🚀
CRUD completo de partidos

Estadísticas de goles, tarjetas y tiempo adicional

Base de datos SQLite embebida

Configuración CORS robusta

Endpoints PATCH para actualizaciones específicas

Requisitos 📋
Go 1.21+

SQLite3

Instalación ⚙️
Clona el repositorio:

bash
Copy
git clone https://github.com/tu-usuario/laliga-api.git
cd laliga-api
Descarga las dependencias:

bash
Copy
go mod download
Ejecución ▶️
bash
Copy
go run main.go
La API estará disponible en http://localhost:8080

Endpoints 🌐
Método	Endpoint	Descripción
GET	/matches	Obtiene todos los partidos
GET	/matches/:id	Obtiene un partido específico
POST	/matches	Crea un nuevo partido
PUT	/matches/:id	Actualiza un partido completo
DELETE	/matches/:id	Elimina un partido
PATCH	/matches/:id/goals	Actualiza los goles
PATCH	/matches/:id/yellowcards	Actualiza tarjetas amarillas
PATCH	/matches/:id/redcards	Actualiza tarjetas rojas
PATCH	/matches/:id/extratime	Actualiza minutos adicionales
Modelo de datos 📊
json
Copy

{
  "match_id": "1",
  "home_team": "Real Madrid",
  "away_team": "Barcelona",
  "date": "2023-10-28",
  "home_goals": 2,
  "away_goals": 1,
  "yellow_cards": 3,
  "red_cards": 1,
  "extra_minutes": 5
}

Ejemplos de uso 💻
Obtener todos los partidos
bash
Copy
curl http://localhost:8080/matches
Crear un partido
bash
Copy
curl -X POST http://localhost:8080/matches \
  -H "Content-Type: application/json" \
  -d '{
    "match_id": "6",
    "home_team": "Real Sociedad",
    "away_team": "Valencia",
    "date": "2023-11-05"
  }'
Actualizar goles
bash
Copy
curl -X PATCH http://localhost:8080/matches/1/goals \
  -H "Content-Type: application/json" \
  -d '{
    "home_goals": 3,
    "away_goals": 2
  }'

Link a la coleccion de postman: https://javier-4375792.postman.co/workspace/Javier's-Workspace~e31ea27b-4d20-4eff-86eb-379f74ab86f3/request/43541136-fe214616-4ef5-4266-b156-87d5def0b012?action=share&creator=43541136&ctx=documentation

Licencia 📄
Este proyecto está bajo la licencia MIT.
