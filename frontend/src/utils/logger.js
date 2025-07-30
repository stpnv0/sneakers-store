const maskToken = (token) => {
  if (!token) return '';
  if (token.length < 10) return '***';
  return `${token.slice(0, 4)}...${token.slice(-4)}`;
};

const isDevelopment = process.env.NODE_ENV === 'development' || !process.env.NODE_ENV;

export const logger = {
  info: (message, data = null) => {
    if (isDevelopment) {
      if (data) {
        // Маскируем конфиденциальные данные
        const sanitizedData = { ...data };
        if (sanitizedData.token) {
          sanitizedData.token = maskToken(sanitizedData.token);
        }
        if (sanitizedData.Authorization) {
          sanitizedData.Authorization = maskToken(sanitizedData.Authorization);
        }
        console.log(message, sanitizedData);
      } else {
        console.log(message);
      }
    }
  },

  error: (message, error = null) => {
    if (error?.response?.headers?.Authorization) {
      error.response.headers.Authorization = maskToken(error.response.headers.Authorization);
    }
    if (error?.config?.headers?.Authorization) {
      error.config.headers.Authorization = maskToken(error.config.headers.Authorization);
    }
    console.error(message, error);
  },

  warn: (message, data = null) => {
    if (isDevelopment) {
      console.warn(message, data);
    }
  }
}; 