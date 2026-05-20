from pydantic_settings import BaseSettings, SettingsConfigDict

# Permet de valider le type de variable automatiquement et de charger les variables d'environnement à partir d'un fichier .env
class Settings(BaseSettings):
    news_api_key: str

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
