# Stage 1: Build the React Application
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy the rest of the source code
COPY . .

# Build the app for production (creates the 'dist' folder)
RUN npm run build

# Stage 2: Serve the App with Nginx
FROM nginx:alpine

# Clean the default Nginx static files
RUN rm -rf /usr/share/nginx/html/*

# Copy the built React files from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80 (standard HTTP port)
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]