# cloud-sql-proxy --port 6000 trusty-charmer-415303:australia-southeast1:three-tier-app-db-d173

export $(cat prod.secrets.env | xargs)

db_pw_encoded=$(python3 -c "import urllib.parse; import os; print(urllib.parse.quote('$POSTGRES_PW', ""))")
conn_string="postgres://postgres:$db_pw_encoded@0.0.0.0:6000/quorum?sslmode=disable"

migrate -source file://migrations -database $conn_string up