import os

def process_value(value):
    if isinstance(value, str):
        return {'type': 'string', 'value': value}
    elif isinstance(value, int):
        return {'type': 'integer', 'value': value}
    elif isinstance(value, float):
        return {'type': 'float', 'value': value}
    elif isinstance(value, datetime.datetime):
        return {'type': 'datetime', 'value': value.isoformat()}
    else:
        return None

def process_file(name):
    directory = os.path.join(os.getcwd(), "src/bi/sql")
    extension = ".sql"

    filename = name + extension
    filepath = os.path.join(directory, filename)

    if os.path.exists(filepath):
        with open(filepath, 'r') as file:
            file_contents = file.read()
            return file_contents
    else:
        return None
