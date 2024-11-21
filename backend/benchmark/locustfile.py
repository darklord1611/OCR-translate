
# /teamspace/studios/this_studio/OCR-translate/backend/data/sample.png
import time
from locust import HttpUser, task, between

def _get_image_part(file_path):
    import os
    file_name = os.path.basename(file_path)
    file_content = open(file_path, 'rb')
    return file_name, file_content


class QuickstartUser(HttpUser):
    wait_time = between(1, 5)

    @task
    def test_upload(self):

        payload = {
            "name": "John Doe",

        }
        
        files = {
            "file": _get_image_part("/teamspace/studios/this_studio/OCR-translate/backend/data/sample.png")
        }

        self.client.post("/upload", data=payload, files=files)