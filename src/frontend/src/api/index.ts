const createBoardUrl = `/api/board/create`

export interface BoardColumn {
    id: string
    text: string
    color: string
}

export interface CreateBoardRequest {
    name: string
    team: string
    owner: string
    columns: BoardColumn[]
}

export interface CreateBoardResponse {
    id: string
}

export const createBoard = async (payload: CreateBoardRequest): Promise<CreateBoardResponse> => {
    try {
        const response = await fetch(createBoardUrl, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
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