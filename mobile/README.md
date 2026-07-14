# Logistics Mobile

Expo/React Native customer app for the Logistics Platform.

## Run

```bash
cp .env.example .env
npm install
npm start
```

When testing on a physical phone, replace `localhost` in `.env` with the local
network address of the computer running the API gateway.

## Checks

```bash
npm run typecheck
npm run format:check
npm run export:android
```

The current UI includes sign-in, registration, password recovery, session
restoration, a demo mode, service booking with device location, map-based live
tracking, order history, wallet, notifications, and profile experiences. API access lives in
`src/api.ts`; authentication tokens are stored with Expo SecureStore.

Screens are separated under `src/screens`, reusable presentation elements live
under `src/components`, server state hooks under `src/hooks`, and backend DTOs
under `src/types.ts`. `App.tsx` only coordinates session state, tabs, and modals.

`eas.json` includes development, preview, production, and store-submission
profiles. Store signing and push credentials must be supplied through the
project owner's Expo, Apple, and Google accounts; they are never committed.
