from flask import Blueprint, request, jsonify
from google.cloud import bigquery
from .utils import process_value, process_file

bi_bp = Blueprint('bi', __name__)

@bi_bp.route('/process_data', methods=['POST'])
def process_data():
    data = request.get_json()
    name = data.get('name')
    template_dict = data.get('template_dict')

    file_contents = process_file(name) if name else None
    if template_dict:
        template = Template(file_contents)
        query = template.substitute(template)

        if file_contents:
            query_job = client.query(query)
            results = query_job.result()

            # Process the query results and construct JSON response
            output = []
            for row in results:
                typed_row = {}
                for key, value in row.items():
                    typed_row[key] = process_value(value)
                output.append(typed_row)

            return jsonify(output), 200
        else:
            return jsonify({'error': 'File does not exist.'}), 404
    return jsonify({'error': 'Invalid template dict provided.'}), 400
