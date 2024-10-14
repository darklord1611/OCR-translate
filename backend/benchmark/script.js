import http from "k6/http";
import { sleep } from "k6";
import { FormData } from "https://jslib.k6.io/formdata/0.0.2/index.js";
import { check } from 'k6';

export const options = {
  vus: 1,
  duration: "30s",
};

const img_temp = open("../data/sample.png", "b");


export default function () {
  // Open the file to upload
  // const binFile = open('./test-file.txt', 'b'); // Open file as binary

  // Prepare multipart form-data
  const data = {
    file: http.file(img_temp, 'sample.png'),  // File to upload
  };

  // Make the POST request to the /upload endpoint
  const res = http.post('http://localhost:8080/upload', data);

  // Verify the response
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
