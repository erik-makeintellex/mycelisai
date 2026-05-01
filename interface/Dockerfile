FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app

# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
RUN apk add --no-cache libc6-compat
COPY interface/package.json interface/package-lock.json* ./interface/
WORKDIR /app/interface
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
ARG MYCELIS_API_HOST=localhost
ARG MYCELIS_API_PORT=8081
ENV MYCELIS_API_HOST=${MYCELIS_API_HOST}
ENV MYCELIS_API_PORT=${MYCELIS_API_PORT}
COPY --from=deps /app/interface/node_modules ./interface/node_modules
COPY interface ./interface
COPY docs ./docs
COPY architecture ./architecture
COPY .state ./.state
COPY README.md ./
WORKDIR /app/interface

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
ENV NEXT_TELEMETRY_DISABLED 1

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1
ARG MYCELIS_API_HOST=localhost
ARG MYCELIS_API_PORT=8081
ENV MYCELIS_API_HOST=${MYCELIS_API_HOST}
ENV MYCELIS_API_PORT=${MYCELIS_API_PORT}

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/interface/public ./interface/public
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/architecture ./architecture
COPY --from=builder /app/.state ./.state
COPY --from=builder /app/README.md ./

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/interface/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/interface/.next/static ./interface/.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
# set hostname to localhost
ENV HOSTNAME "0.0.0.0"

CMD ["node", "interface/server.js"]
