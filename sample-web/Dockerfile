FROM nginx:alpine

# Install envsubst
RUN apk add --no-cache gettext

# Copy static files
COPY index.html /usr/share/nginx/html/index.html.template
COPY entrypoint.sh /usr/share/nginx/html/entrypoint.sh

# Make entrypoint script executable
RUN chmod +x /usr/share/nginx/html/entrypoint.sh

# Use custom entrypoint
ENTRYPOINT ["/usr/share/nginx/html/entrypoint.sh"]
