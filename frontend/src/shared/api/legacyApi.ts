// async function API(url, options = {}) {
//     // If a body exists and it's an object, stringify it
//     if (options.body && typeof options.body === 'object') {
//         options.body = JSON.stringify(options.body);
//     }
//
//     const response = await fetch(baseURL + url, {
//         ...options,
//         headers: {
//             'Content-Type': 'application/json',
//             ...options.headers, // Allow overriding headers if needed
//         },
//     });
//
//     if (!response.ok) {
//         // Helpful for debugging: try to get the error message from the server
//         const errorText = await response.text();
//         throw new Error(`HTTP Error: ${response.status} - ${errorText}`);
//     }
//
//     // Handle empty responses (204 No Content) gracefully
//     if (response.status === 204 || response.headers.get("content-length") === "0") {
//         return null;
//     }
//
//     return response.json();
// }
//
// export default API;

import axios from "axios";

const baseURL = import.meta.env.VITE_API_BASE_URL_PROD;

// const baseURL =
//     process.env.NODE_ENV === "development"
//         ? process.env.REACT_APP_API_BASE_URL_DEV
//         : process.env.REACT_APP_API_BASE_URL_PROD;

const API = axios.create({
  baseURL,
});

export default API;
