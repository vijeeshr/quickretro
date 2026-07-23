import { BoardColumn } from '../models/BoardColumn'

const createBoardUrl = `/api/board/create`

export interface CreateBoardRequest {
  name: string
  team: string
  owner: string
  columns: BoardColumn[]
  cfTurnstileResponse: string
}

export interface CreateBoardResponse {
  id: string
}

export const createBoard = async (payload: CreateBoardRequest): Promise<CreateBoardResponse> => {
  try {
    const response = await fetch(createBoardUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      throw new Error('Network response was not ok')
    }
    const data: CreateBoardResponse = await response.json()
    if (!data.id) {
      throw new Error('Error getting board id from response')
    }

    return data
  } catch (error) {
    console.error('Error:', error)
    throw error // Re-throw the error to maintain the Promise rejection
  }
}

export interface AdminBoardInfo {
  id: string
  name: string
  team: string
  owner: string
  creator: string
  createdAtUtc: number
  autoDeleteAtUtc: number
  ttlSeconds: number
}

export interface AdminBoardsResponse {
  boards: AdminBoardInfo[]
  total: number
  page: number
  limit: number
  totalPages: number
}

export const verifyAdminPasskey = async (passkey: string): Promise<boolean> => {
  const response = await fetch('/api/admin/verify', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Passkey': passkey,
    },
    body: JSON.stringify({ passkey }),
  })
  return response.ok
}

export const getAdminBoards = async (
  passkey: string,
  page: number = 1,
  searchQuery: string = ''
): Promise<AdminBoardsResponse> => {
  const url = `/api/admin/boards?page=${page}&q=${encodeURIComponent(searchQuery)}`
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Passkey': passkey,
    },
  })
  if (!response.ok) {
    throw new Error('Failed to fetch boards')
  }
  return await response.json()
}

export const extendBoardExpiry = async (
  passkey: string,
  boardId: string
): Promise<AdminBoardInfo> => {
  const response = await fetch('/api/admin/board/extend', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Passkey': passkey,
    },
    body: JSON.stringify({ boardId, passkey }),
  })
  if (!response.ok) {
    throw new Error('Failed to extend board expiry')
  }
  return await response.json()
}

export const removeBoardExpiry = async (
  passkey: string,
  boardId: string
): Promise<AdminBoardInfo> => {
  const response = await fetch('/api/admin/board/remove-expiry', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Passkey': passkey,
    },
    body: JSON.stringify({ boardId, passkey }),
  })
  if (!response.ok) {
    throw new Error('Failed to remove board expiry')
  }
  return await response.json()
}

export const deleteBoardByAdmin = async (passkey: string, boardId: string): Promise<void> => {
  const response = await fetch('/api/admin/board/delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Admin-Passkey': passkey,
    },
    body: JSON.stringify({ boardId, passkey }),
  })
  if (!response.ok) {
    throw new Error('Failed to delete board')
  }
}
