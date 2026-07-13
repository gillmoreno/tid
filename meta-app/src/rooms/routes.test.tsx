import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppRoutes } from '../App'
import { resolveSignalingEndpoint } from './api'
import { clearVaultForTests, listRooms } from './vault'

function shapedCreatorPermit(capacity: number): string {
  const payload = btoa(JSON.stringify({
    v: 1,
    purpose: 'create_room',
    jti: 'test-token-id',
    maxMembers: capacity,
    iat: 1,
    exp: 2,
  })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  return `rwp1.${payload}.test-signature`
}

afterEach(async () => {
  cleanup()
  vi.unstubAllGlobals()
  window.sessionStorage.clear()
  window.localStorage.clear()
  await clearVaultForTests()
})

describe('room routes', () => {
  it('resolves the production relative API URL and renders home', () => {
    expect(resolveSignalingEndpoint('/api', 'https://rooms.the-idea-guy.com').href)
      .toBe('https://rooms.the-idea-guy.com/api')
    render(
      <MemoryRouter initialEntries={['/']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    expect(screen.getByRole('button', { name: 'Create a room' })).toBeInTheDocument()
    expect(screen.getByText(/service/i)).toBeInTheDocument()
  })

  it('renders a stable invitation route', () => {
    render(
      <MemoryRouter initialEntries={['/join/invite-public-id']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'A room is waiting.' })).toBeInTheDocument()
    expect(screen.getByText('invite-public-id')).toBeInTheDocument()
  })

  it('keeps the creator permit in the current tab and prefills its capacity', async () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Create a room' }))
    expect(screen.getByRole('form', { name: 'Unlock room creation' })).toBeInTheDocument()
    const permit = shapedCreatorPermit(11)
    fireEvent.change(screen.getByLabelText('Room creator token'), { target: { value: permit } })
    fireEvent.click(screen.getByRole('button', { name: 'Unlock creation' }))

    expect(await screen.findByRole('form', { name: 'Create room' })).toBeInTheDocument()
    expect(screen.getByLabelText('Unique member capacity')).toHaveValue(11)
    expect(screen.getByLabelText('Unique member capacity')).toHaveAttribute('readonly')
    expect(window.sessionStorage.getItem('roomworks.creator-permit')).toBe(permit)
    expect(window.localStorage.length).toBe(0)
  })

  it('clears a creator permit rejected by the API', async () => {
    window.sessionStorage.setItem('roomworks.creator-permit', shapedCreatorPermit(2))
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      error: { code: 'creator_permit_used', message: 'room creator permit has already been used' },
    }), {
      status: 403,
      headers: { 'Content-Type': 'application/json' },
    })))
    render(
      <MemoryRouter initialEntries={['/']}>
        <AppRoutes />
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Create a room' }))
    fireEvent.submit(screen.getByRole('form', { name: 'Create room' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('already been used')
    expect(screen.getByRole('form', { name: 'Unlock room creation' })).toBeInTheDocument()
    expect(window.sessionStorage.getItem('roomworks.creator-permit')).toBeNull()
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
