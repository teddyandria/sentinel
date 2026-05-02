from fastapi import FastAPI
import requests
from datetime import date, timedelta
from .config import settings



app = FastAPI()
@app.get("/")
async def read_root():
    from_date = (date.today() - timedelta(days=7)).isoformat()
    url = (
        f"https://newsapi.org/v2/everything?"
        f"q=Apple&"
        f"from={from_date}&"
        f"sortBy=popularity&"
        f"apiKey={settings.news_api_key}"
    )
    response = requests.get(url)
    return response.json()