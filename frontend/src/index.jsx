import { createRoot } from 'react-dom/client';
import { MaxUI } from '@maxhub/max-ui';
import '@maxhub/max-ui/dist/styles.css';
import { AuthProvider, useRole } from './components/AuthProvider';
import Router from './Router.jsx';
import './styles/index.css';
import './styles/variables.css';

function App() {
  const role = useRole();
  return <Router role={role} />;
}

const Root = () => (
  <MaxUI>
    <AuthProvider>
      <App />
    </AuthProvider>
  </MaxUI>
);

createRoot(document.getElementById('root')).render(<Root />);
