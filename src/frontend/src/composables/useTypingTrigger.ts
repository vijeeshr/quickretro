// import { ref } from "vue";

// export function useTypingIndicator(socket: WebSocket, boardId: string) {
//     const isThrottled = ref(false);
//     let stopTypingTimer: ReturnType<typeof setTimeout> | null = null;

//     const sendTypingEvent = () => {
//         // 1. If we are in the cooldown period, do nothing
//         if (isThrottled.value) return;

//         // // 2. Construct the payload
//         // const request: EventRequest<TypedEvent> = {
//         //     typ: 'typing_start', // Match this with your Go backend switch case
//         //     pyl: { xid: boardId }
//         // };

//         // // 3. Send via WebSocket
//         // socket.send(JSON.stringify(request));

//         // 4. Enter cooldown (Throttle)
//         isThrottled.value = true;
//         setTimeout(() => {
//             isThrottled.value = false;
//         }, 3000); // Only send once every 3 seconds
//     };

//     const handleInput = () => {
//         sendTypingEvent()

//         // Optional: Detect when they STOP typing to clear the status early
//         if (stopTypingTimer) clearTimeout(stopTypingTimer);
//         stopTypingTimer = setTimeout(() => {
//             // Logic for 'typing_stop' if your backend supports it
//             console.log("User stopped typing.");
//         }, 1500);
//     };

//     return { handleInput };
// }


// import { ref } from 'vue'

// export function useTypingTrigger(emit: any, xid: string) {
//     const isThrottled = ref(false)
//     const THROTTLE_MS = 3000 // Adjust based on preference

//     const triggerTyping = () => {
//         if (isThrottled.value) return

//         // 1. Tell the parent to send the WS message
//         emit('typing', xid)

//         // 2. Start cooldown
//         isThrottled.value = true
//         setTimeout(() => {
//             isThrottled.value = false
//         }, THROTTLE_MS)
//     }

//     return { triggerTyping }
// }

import { ref } from 'vue'

export function useTypingTrigger(emit: any) {
    const isThrottled = ref(false)
    const THROTTLE_MS = 3000 // Adjust based on preference

    const triggerTyping = (event?: KeyboardEvent) => {
        // Ignore Enter (save) but allow Shift+Enter (new line)
        if (event?.key === 'Enter' && !event.shiftKey) return
        
        if (isThrottled.value) return

        emit('typing')

        // Start cooldown
        isThrottled.value = true
        setTimeout(() => isThrottled.value = false, THROTTLE_MS)
    }

    return { triggerTyping }
}