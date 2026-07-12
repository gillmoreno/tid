import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppRoutes } from '../App'
import { clearVaultForTests, listRooms } from './vault'

afterEach(async () => {
  cleanup()
  vi.unstubAllGlobals()
  await clearVaultForTests()
})

describe('room routes', () => {
  it('renders a stable invitation route', () => {
    render(
      <MemoryRouter initialEntries={['/join/invite-public-id']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'A room is waiting.' })).toBeInTheDocument()
    expect(screen.getByText('invite-public-id')).toBeInTheDocument()
  })

  it('does not invent a room for an unknown room URL', async () => {
    render(
      <MemoryRouter initialEntries={['/rooms/not-local']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: /not in this device’s vault/i })).toBeInTheDocument()
    expect(screen.getByText(/No placeholder room was created/i)).toBeInTheDocument()
  })

  it('keeps a rejected invitation visible and out of IndexedDB', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      code: 'invalid_invite',
      message: 'That invitation key is not valid.',
    }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    })))
    render(
      <MemoryRouter initialEntries={['/join/rejected-invite']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    fireEvent.change(screen.getByLabelText('Invitation package'), { target: { value: 'bad-key' } })
    fireEvent.submit(screen.getByRole('button', { name: 'Join encrypted room' }).closest('form')!)

    expect(await screen.findByRole('alert')).toHaveTextContent('This invitation could not be verified.')
    expect(screen.getByText('rejected-invite')).toBeInTheDocument()
    await waitFor(async () => expect(await listRooms()).toEqual([]))
  })
})
