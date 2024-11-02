import { Auth0Provider } from '@auth0/auth0-react';
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

const root = createRoot(document.getElementById('root')!);
console.log("Root: ", window.location.origin)
root.render(
  <Auth0Provider
    domain="dev-bxn245l6be2yzhil.us.auth0.com"
    clientId="Ne5QmRSrbFuXW9p0ahbQUrIETB6lWhQL"
    authorizationParams={{
      redirect_uri: window.location.origin,
      audience: "creeper-keeper-resource",
      scope: "read:all write:all",
    }}
  >
    <App />
  </Auth0Provider>,
)


