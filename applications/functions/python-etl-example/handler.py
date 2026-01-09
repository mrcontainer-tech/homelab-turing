import json
import sys
from datetime import datetime
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "healthy"}), 200


@app.route('/', methods=['POST'])
def handle():
    """
    Knative Python ETL Function
    
    This function demonstrates a simple ETL pipeline:
    1. Extract: Read data from the request
    2. Transform: Process and clean the data
    3. Load: Output the transformed data (or write to DB/S3)
    
    Returns:
        JSON response with transformed data
    """
    
    try:
        # Extract: Parse incoming JSON data
        body = request.get_json(force=True)
        
        # Transform: Process the data
        transformed_data = transform_data(body)
        
        # Load: In a real scenario, you would:
        # - Write to a database (PostgreSQL, MongoDB, etc.)
        # - Upload to S3/MinIO
        # - Send to another API
        # - Publish to a message queue
        
        result = {
            "status": "success",
            "timestamp": datetime.utcnow().isoformat(),
            "records_processed": len(transformed_data.get("items", [])),
            "data": transformed_data
        }
        
        return jsonify(result), 200
    
    except json.JSONDecodeError as e:
        return jsonify({
            "status": "error",
            "message": f"Invalid JSON: {str(e)}"
        }), 400
    
    except Exception as e:
        return jsonify({
            "status": "error",
            "message": f"Processing error: {str(e)}"
        }), 500


def transform_data(data):
    """
    Transform the input data according to business rules
    
    Example transformations:
    - Normalize field names
    - Clean data
    - Aggregate values
    - Filter records
    """
    
    transformed = {
        "items": [],
        "metadata": {
            "transformed_at": datetime.utcnow().isoformat(),
            "version": "1.0"
        }
    }
    
    # Example: If input has a list of items, process each
    if "items" in data:
        for item in data["items"]:
            # Clean and transform each item
            cleaned_item = {
                "id": item.get("id", "unknown"),
                "name": item.get("name", "").strip().upper(),
                "value": float(item.get("value", 0)),
                "processed": True
            }
            transformed["items"].append(cleaned_item)
    
    # Example: Handle single object input
    elif "name" in data or "value" in data:
        transformed["items"].append({
            "id": data.get("id", "single"),
            "name": data.get("name", "").strip().upper(),
            "value": float(data.get("value", 0)),
            "processed": True
        })
    
    return transformed


if __name__ == "__main__":
    # Run Flask app on port 8080 (Knative requirement)
    app.run(host='0.0.0.0', port=8080)

