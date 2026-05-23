import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';

export default function WorkspaceShell() {
  return (
    <div className="min-h-screen bg-[#FFFDF7]">
      <Navbar />
      <div className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </div>
    </div>
  );
}
