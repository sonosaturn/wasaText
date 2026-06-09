import axios from "axios";

const instance = axios.create({
	baseURL: `http://${window.location.hostname}:3000`,
	timeout: 1000 * 5
});

export default instance;
