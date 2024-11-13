from sentence_transformers import SentenceTransformer
import pandas as pd
import json

# Load the model
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

# Load the CSV file
csv_file_path = 'dataset.csv'
df = pd.read_csv(csv_file_path)

# Process the data and generate embeddings
data = []
for review in df['review']:
    embedding = model.encode(review).tolist()  # Convert to list to make JSON serializable
    data.append({'review': review, 'embedding': embedding})

# Save to JSON file
json_file_path = 'dataset.json'
with open(json_file_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print(f"Data has been written to {json_file_path}")
