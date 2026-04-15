import { NavLink, Outlet } from 'react-router-dom'

export function Layout() {
  return (
    <div className="flex min-h-screen bg-gray-950 text-gray-100">
      <aside className="w-56 shrink-0 border-r border-gray-800 flex flex-col px-4 py-6 gap-1">
        <span className="text-xs font-semibold tracking-widest text-gray-500 uppercase mb-4">
          Monitoring
        </span>
        <NavItem to="/runs">Runs</NavItem>
      </aside>
      <main className="flex-1 overflow-auto p-8">
        <Outlet />
      </main>
    </div>
  )
}

function NavItem({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `block px-3 py-2 rounded-md text-sm font-medium transition-colors ${
          isActive
            ? 'bg-gray-800 text-white'
            : 'text-gray-400 hover:text-white hover:bg-gray-800/60'
        }`
      }
    >
      {children}
    </NavLink>
  )
}
