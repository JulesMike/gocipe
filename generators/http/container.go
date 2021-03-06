package http

var tmplContainer = `
const (
	//EnvironmentProd represents production environment
	EnvironmentProd = "PROD"
	//EnvironmentDev represents development environment
	EnvironmentDev  = "DEV"
)

var (
	env string
	port string
	dsn string
	db  *sql.DB
)

func init() {
	var (
		err error
	)

	godotenv.Load()

	env  = os.Getenv("ENV")
	port = os.Getenv("PORT")
	dsn  = os.Getenv("DSN")

	if env == "" {
		log.Fatal("Environment variable ENV must be defined. Possible values are: DEV PROD")
	}

	if port == "" {
		port = "8888"
	}

	if dsn == "" {
		log.Fatal("Environment variable DSN must be defined. Example: postgres://user:pass@host/db?sslmode=disable")
	}

	db, err = sql.Open("postgres", dsn)
	if err == nil {
		log.Println("Connected to database successfully.")
	} else if (env == EnvironmentDev) {
		log.Println("Database connection failed: ", err)
	} else {
		log.Fatal("Database connection failed: ", err)
	}

	err = db.Ping()
	if err == nil {
		log.Println("Pinged database successfully.")
	} else if (env == EnvironmentDev) {
		log.Println("Database ping failed: ", err)
	} else {
		log.Fatal("Database ping failed: ", err)
	}
}
`

//GenerateContainer generates code to load configuration for the application
func GenerateContainer() (string, error) {
	return tmplContainer, nil
}
