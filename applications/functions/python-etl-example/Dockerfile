FROM python:3.11-slim

WORKDIR /app

# Copy requirements and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY handler.py .

# Expose port 8080 (Knative requirement)
EXPOSE 8080

# Run the Flask application
CMD ["python", "handler.py"]
