import os

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
