import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';

import { FormaSessionProvider } from '@/hooks/use-forma-session';
import { AppRouter } from '@/routes';

import '@/styles/tokens.css';
import '@/components/shell.css';

const root = document.getElementById('root');
if (!root) {
  throw new Error('Root element #root not found');
}

ReactDOM.createRoot(root).render(
  <React.StrictMode>
    <BrowserRouter>
      <FormaSessionProvider>
        <AppRouter />
      </FormaSessionProvider>
    </BrowserRouter>
  </React.StrictMode>,
);
