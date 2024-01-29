const createBoardUrl = `http://localhost:8080/api/board/create`

export interface CreateBoardResponse {
    id: string
}
  
export const createBoard = async (name: string, team: string, user: string): Promise<CreateBoardResponse> => {
    try {
        const response = await fetch(createBoardUrl, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*", // TEMPORARY
            },
            body: JSON.stringify({ name, team, owner: user }),            
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