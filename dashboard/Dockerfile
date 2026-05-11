FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY pipeline ./pipeline
COPY dashboard/app.py ./app.py
CMD ["streamlit", "run", "app.py", "--server.address=0.0.0.0", "--server.port=8501"]
