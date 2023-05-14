import os
import pandas as pd
from google.cloud import tasks_v2
from google.protobuf import timestamp_pb2
from datetime import datetime, timedelta
from abc import ABC, abstractmethod


class CloudTasksBatchProcessor(BatchProcessor):
    def __init__(self, folder_path, chunk_size=10):
        super().__init__(folder_path, chunk_size)
        self.client = tasks_v2.CloudTasksClient()

    def process_files(self):
        # get all text files in the directory
        files = [f for f in os.listdir(self.folder_path) if f.endswith('.txt')]

        # iterate over the files and read them in chunks
        for file in files:
            with open(os.path.join(self.folder_path, file), 'r') as f:
                # save the first line as the rate limit for this site
                self.rate_limits[file] = int(f.readline().strip())
                # initialize the last scheduled time to now
                self.last_scheduled_times[file] = datetime.utcnow()

                # parent is the queue that will process the task
                parent = self.client.queue_path('project_id', 'location_id', 'queue_id')

                # now process the rest of the file in chunks
                for piece in self.read_in_chunks(f):
                    df = pd.DataFrame(piece, columns=['Product_URL'])
                    for index, row in df.iterrows():
                        url = row['Product_URL']

                        # construct the request body
                        task = {
                            'app_engine_http_request': {  # Specify the type of request.
                                'http_method': 'POST',
                                'relative_uri': '/example_task_handler',  # Specify the handler.
                                'body': url.encode(),  # Send the URL in the body.
                                'headers': {
                                    'Content-Type': 'text/plain',
                                },
                            },
                        }

                        # Convert "seconds from now" into an rfc3339 datetime string.
                        d = self.last_scheduled_times[file] + timedelta(seconds=1/self.rate_limits[file])
                        timestamp = timestamp_pb2.Timestamp()
                        timestamp.FromDatetime(d)

                        # Add the timestamp to the tasks.
                        task['schedule_time'] = timestamp

                        # Use the client to build and send the task.
                        response = self.client.create_task(parent, task)
                        print('Created task {}'.format(response.name))

                        # Update the last scheduled time
                        self.last_scheduled_times[file] = d