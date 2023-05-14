from google.cloud import storage


def read_file_from_gcs_bucket(bucket_name, file_name, chunk_size=1024):
    client = storage.Client()
    bucket = client.get_bucket(bucket_name)
    blob = bucket.blob(file_name)

    with blob.open("rb") as file:
        remainder = b""
        while True:
            chunk = file.read(chunk_size)
            if not chunk:
                break

            lines = (remainder + chunk).splitlines(keepends=True)
            if lines:
                if lines[-1].endswith(b"\n"):
                    yield lines.pop(0)
                remainder = lines.pop()

            for line in lines:
                yield line

        if remainder:
            yield remainder

def read_file_local(file_path, chunk_size=1024):
    with open(file_path, "rb") as file:
        remainder = b""
        while True:
            chunk = file.read(chunk_size)
            if not chunk:
                break

            lines = (remainder + chunk).splitlines(keepends=True)
            if lines:
                if lines[-1].endswith(b"\n"):
                    yield lines.pop(0)
                remainder = lines.pop()

            for line in lines:
                yield line

        if remainder:
            yield remainder