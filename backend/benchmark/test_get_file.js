import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  // Key configurations for Stress in this section
  vus: 5,
  duration: "30s"
};

export default () => {
  const urlRes = http.get('http://localhost:8080/api/v1/app/translate');
  sleep(1);
  // MORE STEPS
  // Here you can have more steps or complex script
  // Step1
  // Step2
  // etc.
};