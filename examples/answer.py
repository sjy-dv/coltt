from sentence_transformers import SentenceTransformer
import pandas as pd
import json
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')


data = []

q1 = 'Where is a place with a comfortable indoor environment?'
q2 = 'Where is a good restaurant to go to with family?'
q3 = 'A restaurant I strongly do not recommend.'
data.append({'review': q1, 'embedding':model.encode(q1).tolist()})
data.append({'review': q2, 'embedding':model.encode(q2).tolist()})
data.append({'review': q3, 'embedding':model.encode(q3).tolist()})

json_file_path = 'review_question.json'
with open(json_file_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)