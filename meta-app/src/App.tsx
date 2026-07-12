import { Navigate, Route, Routes } from 'react-router-dom'
import { HomePage } from './rooms/HomePage'
import { JoinPage } from './rooms/JoinPage'
import { RoomPage } from './rooms/RoomPage'
import { Shell } from './rooms/Shell'

export function AppRoutes() {
  return (
    <Shell>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/rooms/:roomId" element={<RoomPage />} />
        <Route path="/join/:inviteId" element={<JoinPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Shell>
  )
}

export default AppRoutes
