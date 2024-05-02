import axios from 'axios';
import configs from '../../../dev.configs.json';

export const axiosClient = axios.create({
    baseURL: configs.backendBaseUrl[0],
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
})

export const privateClient = axios.create({
    baseURL: configs.backendBaseUrl[1],
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'Authorization': 'Basic ZGV2OmRldg=='
    }
})

















