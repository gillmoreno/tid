import { useEffect, useMemo, useRef } from 'react'
import { disconnectBridgeSource, handleBridgeMessage, type RoomBridge } from './bridge'
import {
  incrementCounter,
  renderCounterBundle,
  roomState,
  subscribeToRoom,
  type CodeBundle,
} from './roomService'
import type { VaultRoom } from './types'

type RoomFrameProps = {
  room: VaultRoom
  bundle: CodeBundle
}

export function RoomFrame({ room, bundle }: RoomFrameProps) {
  const frameRef = useRef<HTMLIFrameElement>(null)
  const rendered = useMemo(() => renderCounterBundle(bundle), [bundle])

  useEffect(() => {
    const source = frameRef.current?.contentWindow
    if (!source) return
    const bridge: RoomBridge = {
      getState: () => roomState(room),
      update: () => incrementCounter(room),
      subscribe: (listener) => subscribeToRoom(room.id, listener),
    }
    const listener = (event: MessageEvent<unknown>) => {
      void handleBridgeMessage(event, source, rendered.nonce, bridge)
    }
    window.addEventListener('message', listener)
    return () => {
      disconnectBridgeSource(source)
      window.removeEventListener('message', listener)
    }
  }, [rendered.nonce, room])

  return (
    <iframe
      ref={frameRef}
      className="room-frame"
      title={`${room.title} application`}
      sandbox="allow-scripts"
      referrerPolicy="no-referrer"
      src={rendered.src}
    />
  )
}
