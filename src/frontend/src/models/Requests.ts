import { BoardColumn } from "../api"
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
    nickname: string
    grp: string
    msg: string
    cat: string
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
}

export interface UserClosingResponse {
    typ: 'closing'
    users: OnlineUser[]
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
    nickname: string
    msg: string
    cat: string
    likes: string
    liked: boolean
    mine: boolean
}

export interface LikeMessageResponse {
    typ: 'like'
    id: string
    likes: string
    liked: boolean
}

export interface DeleteMessageResponse {
    typ: 'del'
    id: string
}

export type SocketResponse = RegisterResponse | MaskResponse | LockResponse | MessageResponse |
    LikeMessageResponse | DeleteMessageResponse | UserClosingResponse

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
            case 'closing':
                return json as UserClosingResponse
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