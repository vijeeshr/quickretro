import { BoardColumn } from './BoardColumn'
import { OnlineUser } from './OnlineUser'

export interface EventRequest<T> {
  typ: string
  pyl: T
}

export type RegisterEvent = Record<string, never>

export interface MaskEvent {
  mask: boolean
}

export interface LockEvent {
  lock: boolean
}

export interface SaveMessageEvent {
  id: string
  nickname: string
  msg: string
  cat: string
  anon: boolean
  pid: string
}

export interface LikeMessageEvent {
  msgId: string
  like: boolean
}

export interface DeleteMessageEvent {
  msgId: string
  commentIds: string[]
}

export type DeleteAllEvent = Record<string, never>

export interface CategoryChangeEvent {
  msgId: string
  commentIds: string[]
  newcat: string
  oldcat: string
}

export interface TimerEvent {
  expiryDurationInSeconds: number
  stop: boolean
}

export interface ColumnsChangeEvent {
  columns: BoardColumn[]
}

export type TypedEvent = Record<string, never>

export interface RegisterResponse {
  typ: 'reg'
  boardName: string
  boardTeam: string
  columns: BoardColumn[]
  boardStatus: string
  xid: string
  boardMasking: boolean
  boardLock: boolean
  isBoardOwner: boolean
  mine: boolean
  users: OnlineUser[]
  messages: MessageResponse[]
  comments: MessageResponse[]
  timerExpiresInSeconds: number
  boardExpiryUtcSeconds: number // Unix Timestamp Seconds
  notifyNewBoardExpiry: boolean
}

export interface UserJoiningResponse {
  typ: 'joining'
  nickname: string
  xid: string
}

export interface UserClosingResponse {
  typ: 'closing'
  xid: string
}

export interface MaskResponse {
  typ: 'mask'
  mask: boolean
}

export interface LockResponse {
  typ: 'lock'
  lock: boolean
}

export interface MessageResponse {
  typ: 'msg'
  id: string
  byxid: string
  nickname: string
  msg: string
  cat: string
  likes: number
  liked: boolean
  mine: boolean
  anon: boolean
  pid: string
}

export interface LikeMessageResponse {
  typ: 'like'
  id: string
  likes: number
  liked: boolean
}

export interface DeleteMessageResponse {
  typ: 'del'
  id: string
}

export interface DeleteAllResponse {
  typ: 'delall'
}

export interface CategoryChangeResponse {
  typ: 'catchng'
  id: string
  newcat: string
}

export interface TimerResponse {
  typ: 'timer'
  expiresInSeconds: number
}

export interface ColumnsChangeResponse {
  typ: 'colreset'
  columns: BoardColumn[]
}

export interface TypedResponse {
  typ: 't'
  xid: string
}

export type SocketResponse =
  | RegisterResponse
  | MaskResponse
  | LockResponse
  | MessageResponse
  | LikeMessageResponse
  | DeleteMessageResponse
  | DeleteAllResponse
  | CategoryChangeResponse
  | UserJoiningResponse
  | UserClosingResponse
  | TimerResponse
  | ColumnsChangeResponse
  | TypedResponse

export function toSocketResponse(json: unknown): SocketResponse | null {
  const obj = json as Record<string, unknown>
  if (obj && obj.typ) {
    switch (obj.typ) {
      case 'reg':
        return obj as unknown as RegisterResponse
      case 'mask':
        return obj as unknown as MaskResponse
      case 'lock':
        return obj as unknown as LockResponse
      case 'msg':
        return obj as unknown as MessageResponse
      case 'like':
        return obj as unknown as LikeMessageResponse
      case 'del':
        return obj as unknown as DeleteMessageResponse
      case 'delall':
        return obj as unknown as DeleteAllResponse
      case 'catchng':
        return obj as unknown as CategoryChangeResponse
      case 'joining':
        return obj as unknown as UserJoiningResponse
      case 'closing':
        return obj as unknown as UserClosingResponse
      case 'timer':
        return obj as unknown as TimerResponse
      case 'colreset':
        return obj as unknown as ColumnsChangeResponse
      case 't':
        return obj as unknown as TypedResponse
      // const data: MaskResponse = json
      // return data

      // return {
      //     typ: 'reg',
      //     boardName: json.boardName,
      //     boardTeam: json.boardTeam,
      //     boardStatus: json.boardStatus,
      //     boardMasking: json.boardMasking,
      //     isBoardOwner: json.isBoardOwner,
      //     user: json.user as UserDetails,
      // } as RegisterResponse
      default:
        return null // Handle unknown "typ" values as needed
    }
  }

  return null
}
