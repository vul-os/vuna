from flask import Blueprint, request, jsonify
from string import Template

from google.cloud import bigquery

from src.bi.utils import process_file


client = bigquery.Client()

bi_bp = Blueprint('bi', __name__)

@bi_bp.route('/process_data', methods=['POST'])
def process_data():
    data = request.get_json()
    print(data)
    name = data.get('name')
    template_dict = data.get('template_dict')

    file_contents = process_file(name) if name else None
    print(file_contents)
    if template_dict:
        template = Template(file_contents)
        query = template.substitute(template_dict)
        print(query)
        if file_contents:
            query_job = client.query(query)
            results = query_job.result()
            return jsonify(results), 200
        else:
            return jsonify({'error': 'File does not exist.'}), 404
    return jsonify({'error': 'Invalid template dict provided.'}), 400
