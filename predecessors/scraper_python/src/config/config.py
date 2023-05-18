import os
import toml

def parse_config():
    local_file_path = os.path.join(os.path.dirname(__file__), 'config.toml')

    config_string = ""
    if os.path.exists(local_file_path):
        with open(local_file_path, 'r') as file:
            config_string = file.read()
    else:
        config_string = os.environ.get('CONFIG')

    config = toml.loads(config_string)
    return config