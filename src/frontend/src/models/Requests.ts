import { BoardColumn } from "./BoardColumn"
import { OnlineUser } from "./OnlineUser"

export interface EventRequest<T> {
    typ: string
    pyl: T
}

export interface RegisterEvent {
    by: string
    nickname: string
    xid: string
    grp: string
}

export interface MaskEvent {
    by: string
    grp: string
    mask: boolean
}

export interface LockEvent {
    by: string
    grp: string
    lock: boolean
}

export interface SaveMessageEvent {
    id: string
    by: string
    byxid: string
    nickname: string
    grp: string
    msg: string
    cat: string
    anon: boolean
    pid: string
}

export interface LikeMessageEvent {
    msgId: string
    by: string
    like: boolean
}

export interface DeleteMessageEvent {
    msgId: string
    by: string
    grp: string
    commentIds: string[]
}

export interface DeleteAllEvent {
    by: string
    grp: string
}

export interface CategoryChangeEvent {
    msgId: string
    commentIds: string[]
    by: string
    grp: string
    newcat: string
    oldcat: string
}

export interface TimerEvent {
    by: string
    grp: string
    expiryDurationInSeconds: number
    stop: boolean
}

export interface ColumnsChangeEvent {
    by: string
    grp: string
    columns: BoardColumn[]
}

export interface RegisterResponse {
    typ: 'reg'
    boardName: string
    boardTeam: string
    columns: BoardColumn[]
    boardStatus: string
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

export type SocketResponse = RegisterResponse | MaskResponse | LockResponse | MessageResponse |
    LikeMessageResponse | DeleteMessageResponse | DeleteAllResponse | CategoryChangeResponse | UserJoiningResponse | UserClosingResponse | TimerResponse | ColumnsChangeResponse

export function toSocketResponse(json: any): SocketResponse | null {

    if (json && json.typ) {
        switch (json.typ) {
            case 'reg':
                return json as RegisterResponse
            case 'mask':
                return json as MaskResponse
            case 'lock':
                return json as LockResponse    
            case 'msg':
                return json as MessageResponse
            case 'like':
                return json as LikeMessageResponse
            case 'del':
                return json as DeleteMessageResponse
            case 'delall':
                return json as DeleteAllResponse
            case 'catchng':
                return json as CategoryChangeResponse
            case 'joining':
                return json as UserJoiningResponse
            case 'closing':
                return json as UserClosingResponse
            case 'timer':
                return json as TimerResponse
            case 'colreset':
                return json as ColumnsChangeResponse
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