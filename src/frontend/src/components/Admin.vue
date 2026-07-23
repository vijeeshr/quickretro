<template>
  <div class="min-h-screen bg-slate-900 text-slate-100 flex flex-col">
    <!-- Header -->
    <header
      class="border-b border-slate-800 bg-slate-950 px-6 py-4 flex items-center justify-between shadow-md"
    >
      <div class="flex items-center space-x-3">
        <a
          href="/"
          class="flex items-center space-x-2 text-xl font-bold tracking-tight text-white hover:opacity-90 transition"
        >
          <span class="bg-linear-to-r from-blue-500 to-indigo-500 text-transparent bg-clip-text"
            >QuickRetro</span
          >
          <span
            class="text-xs uppercase bg-blue-900/60 text-blue-300 font-semibold px-2 py-0.5 rounded border border-blue-700/50"
            >Admin</span
          >
        </a>
      </div>
      <div v-if="isAuthenticated" class="flex items-center space-x-4">
        <button
          class="text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 px-3 py-1.5 rounded-lg border border-slate-700 transition flex items-center space-x-1"
          @click="logout"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
            />
          </svg>
          <span>Lock / Exit</span>
        </button>
      </div>
    </header>

    <!-- Passkey Login View -->
    <div v-if="!isAuthenticated" class="flex-1 flex items-center justify-center p-4">
      <div class="w-full max-w-md bg-slate-950 border border-slate-800 rounded-2xl p-8 shadow-2xl">
        <div class="text-center mb-6">
          <div
            class="mx-auto w-12 h-12 bg-blue-600/20 text-blue-400 rounded-xl flex items-center justify-center mb-3"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-white">Admin Authentication</h2>
          <p class="text-slate-400 text-sm mt-1">
            Enter ADMIN_PASSKEY to access board administration.
          </p>
        </div>

        <form class="space-y-4" @submit.prevent="handleLogin">
          <div>
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2"
              >Admin Passkey</label
            >
            <input
              v-model="inputPasskey"
              type="password"
              placeholder="••••••••••••"
              class="w-full bg-slate-900 border border-slate-700 focus:border-blue-500 rounded-lg px-4 py-2.5 text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition"
              required
              autofocus
            />
          </div>

          <div
            v-if="loginError"
            class="p-3 bg-red-950/60 border border-red-800/60 text-red-300 rounded-lg text-sm flex items-center space-x-2"
          >
            <svg
              class="w-4 h-4 shrink-0 text-red-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>{{ loginError }}</span>
          </div>

          <button
            type="submit"
            :disabled="isAuthenticating"
            class="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2.5 rounded-lg transition shadow-lg shadow-blue-600/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
          >
            <span
              v-if="isAuthenticating"
              class="inline-block animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full"
            ></span>
            <span>{{ isAuthenticating ? 'Authenticating...' : 'Unlock Admin Portal' }}</span>
          </button>
        </form>
      </div>
    </div>

    <!-- Admin Dashboard View -->
    <main v-else class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">
      <!-- Title & Summary -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-white">Board Management</h1>
          <p class="text-sm text-slate-400 mt-1">
            Overview of registered retrospective boards and expiry controls.
          </p>
        </div>
        <div
          class="flex flex-col sm:flex-row items-stretch sm:items-center space-y-3 sm:space-y-0 sm:space-x-3"
        >
          <!-- Search Input -->
          <div class="relative w-full sm:w-72">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search name, team, or ID..."
              class="w-full bg-slate-900 border border-slate-700 focus:border-blue-500 rounded-lg pl-9 pr-8 py-1.5 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition"
              @input="onSearchInput"
            />
            <svg
              class="w-4 h-4 text-slate-500 absolute left-3 top-2.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <button
              v-if="searchQuery"
              @click="clearSearch"
              class="absolute right-2.5 top-2 text-slate-500 hover:text-slate-300 transition"
              title="Clear search"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <span
            class="bg-slate-800 border border-slate-700 text-slate-300 px-3 py-1.5 rounded-lg text-sm font-medium whitespace-nowrap text-center"
          >
            Total Boards: <strong class="text-white">{{ totalBoards }}</strong>
          </span>
          <button
            :disabled="isLoading"
            class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg border border-slate-700 transition flex items-center justify-center"
            title="Refresh list"
            @click="fetchBoards(currentPage)"
          >
            <svg
              class="w-4 h-4"
              :class="{ 'animate-spin': isLoading }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>
        </div>
      </div>

      <!-- Action Toast Message -->
      <div
        v-if="toastMessage"
        class="p-4 bg-emerald-950/80 border border-emerald-700 text-emerald-200 rounded-xl text-sm flex items-center justify-between shadow-lg"
      >
        <div class="flex items-center space-x-2">
          <svg
            class="w-5 h-5 text-emerald-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M5 13l4 4L19 7"
            />
          </svg>
          <span>{{ toastMessage }}</span>
        </div>
        <button class="text-emerald-400 hover:text-emerald-200" @click="toastMessage = ''">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <!-- Boards Table Card -->
      <div class="bg-slate-950 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr
                class="bg-slate-900/80 border-b border-slate-800 text-xs font-semibold uppercase tracking-wider text-slate-400"
              >
                <th class="py-3.5 px-4">Board Name</th>
                <th class="py-3.5 px-4">Team</th>
                <th class="py-3.5 px-4">Created At</th>
                <th class="py-3.5 px-4">Expiry Date</th>
                <th class="py-3.5 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-sm">
              <tr v-if="isLoading" class="text-slate-400 text-center py-8">
                <td colspan="5" class="py-12">
                  <div class="flex items-center justify-center space-x-2 text-slate-400">
                    <span
                      class="inline-block animate-spin w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full"
                    ></span>
                    <span>Loading boards...</span>
                  </div>
                </td>
              </tr>

              <tr v-else-if="boards.length === 0" class="text-slate-400 text-center py-8">
                <td colspan="5" class="py-12">
                  <p class="text-slate-400 font-medium">No boards found.</p>
                  <p class="text-xs text-slate-500 mt-1">
                    Newly created boards will be listed here.
                  </p>
                </td>
              </tr>

              <tr v-for="board in boards" :key="board.id" class="hover:bg-slate-900/50 transition">
                <!-- Board Name (Hyperlink) -->
                <td class="py-3.5 px-4 font-medium text-white">
                  <a
                    :href="`/board/${board.id}`"
                    target="_blank"
                    class="text-blue-400 hover:text-blue-300 hover:underline items-center space-x-1.5 group inline-flex"
                  >
                    <span>{{ board.name || 'Untitled Board' }}</span>
                    <svg
                      class="w-3.5 h-3.5 text-slate-500 group-hover:text-blue-300 transition"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                      />
                    </svg>
                  </a>
                  <div class="text-xs text-slate-500 font-mono mt-0.5">ID: {{ board.id }}</div>
                </td>

                <!-- Team -->
                <td class="py-3.5 px-4 text-slate-300">
                  {{ board.team || '-' }}
                </td>

                <!-- Created At -->
                <td class="py-3.5 px-4 text-slate-300 whitespace-nowrap">
                  {{ formatDate(board.createdAtUtc) }}
                </td>

                <!-- Expiry Date -->
                <td class="py-3.5 px-4 whitespace-nowrap">
                  <div
                    v-if="board.autoDeleteAtUtc <= 0 || board.ttlSeconds === -1"
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-950/80 text-emerald-300 border border-emerald-700/60"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 mr-1.5"></span>
                    No Expiry (Persistent)
                  </div>
                  <div v-else class="space-y-1">
                    <div class="text-slate-200 text-xs font-medium">
                      {{ formatDate(board.autoDeleteAtUtc) }}
                    </div>
                    <div
                      class="inline-flex items-center px-2 py-0.5 rounded text-xs bg-amber-950/60 text-amber-300 border border-amber-800/50"
                    >
                      {{ formatTTL(board.ttlSeconds) }}
                    </div>
                  </div>
                </td>

                <!-- Actions -->
                <td class="py-3.5 px-4 text-right whitespace-nowrap">
                  <div class="inline-flex items-center space-x-2">
                    <!-- Extend Expiry Button -->
                    <button
                      :disabled="actionInProgress === board.id"
                      class="px-2.5 py-1.5 bg-blue-600/20 hover:bg-blue-600 text-blue-300 hover:text-white rounded-lg text-xs font-medium border border-blue-500/30 hover:border-transparent transition disabled:opacity-50"
                      title="Extend board expiry duration"
                      @click="handleExtendExpiry(board.id)"
                    >
                      Extend Expiry
                    </button>

                    <!-- Remove Expiry Button -->
                    <button
                      :disabled="actionInProgress === board.id || board.autoDeleteAtUtc <= 0"
                      class="px-2.5 py-1.5 bg-emerald-600/20 hover:bg-emerald-600 text-emerald-300 hover:text-white rounded-lg text-xs font-medium border border-emerald-500/30 hover:border-transparent transition disabled:opacity-40 disabled:cursor-not-allowed"
                      title="Make board persistent (remove expiry)"
                      @click="handleRemoveExpiry(board.id)"
                    >
                      Remove Expiry
                    </button>

                    <!-- Delete Button -->
                    <button
                      :disabled="actionInProgress === board.id"
                      class="px-2.5 py-1.5 bg-red-600/20 hover:bg-red-600 text-red-300 hover:text-white rounded-lg text-xs font-medium border border-red-500/30 hover:border-transparent transition disabled:opacity-50"
                      title="Permanently delete board"
                      @click="confirmDelete(board)"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Controls -->
        <div
          class="bg-slate-900/90 border-t border-slate-800 px-6 py-4 flex flex-col sm:flex-row items-center justify-between gap-4"
        >
          <div class="text-xs text-slate-400">
            Showing <span class="font-semibold text-slate-200">{{ showingStart }}</span> to
            <span class="font-semibold text-slate-200">{{ showingEnd }}</span> of
            <span class="font-semibold text-slate-200">{{ totalBoards }}</span> boards
          </div>

          <div class="flex items-center space-x-2">
            <button
              :disabled="currentPage <= 1 || isLoading"
              class="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-200 text-xs font-medium rounded-lg border border-slate-700 transition"
              @click="changePage(currentPage - 1)"
            >
              Previous
            </button>

            <span class="text-xs font-medium text-slate-300 px-2">
              Page {{ currentPage }} of {{ totalPages || 1 }}
            </span>

            <button
              :disabled="currentPage >= totalPages || isLoading || totalPages === 0"
              class="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-slate-200 text-xs font-medium rounded-lg border border-slate-700 transition"
              @click="changePage(currentPage + 1)"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Delete Confirmation Modal -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div
        class="w-full max-w-md bg-slate-950 border border-slate-800 rounded-2xl p-6 shadow-2xl space-y-4"
      >
        <div class="flex items-center space-x-3 text-red-400">
          <div class="w-10 h-10 bg-red-600/20 rounded-xl flex items-center justify-center">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </div>
          <h3 class="text-lg font-bold text-white">Delete Board</h3>
        </div>

        <p class="text-slate-300 text-sm">
          Are you sure you want to permanently delete board
          <strong class="text-white">'{{ deleteTarget.name }}'</strong> (ID:
          <code class="text-xs bg-slate-800 px-1 py-0.5 rounded">{{ deleteTarget.id }}</code
          >)? All cards, comments, and presence data will be removed.
        </p>

        <div class="flex items-center justify-end space-x-3 pt-2">
          <button
            class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-medium rounded-lg transition"
            @click="deleteTarget = null"
          >
            Cancel
          </button>
          <button
            :disabled="actionInProgress === deleteTarget.id"
            class="px-4 py-2 bg-red-600 hover:bg-red-500 text-white text-sm font-medium rounded-lg transition shadow-lg shadow-red-600/20 disabled:opacity-50"
            @click="executeDelete"
          >
            {{ actionInProgress === deleteTarget.id ? 'Deleting...' : 'Confirm Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  AdminBoardInfo,
  getAdminBoards,
  extendBoardExpiry,
  removeBoardExpiry,
  deleteBoardByAdmin,
  verifyAdminPasskey,
} from '../api'

const passkeyKey = 'quickretro_admin_passkey'

const passkey = ref<string>(sessionStorage.getItem(passkeyKey) || '')
const inputPasskey = ref<string>('')
const isAuthenticated = ref<boolean>(false)
const isAuthenticating = ref<boolean>(false)
const loginError = ref<string>('')

const boards = ref<AdminBoardInfo[]>([])
const totalBoards = ref<number>(0)
const currentPage = ref<number>(1)
const totalPages = ref<number>(1)
const isLoading = ref<boolean>(false)
const actionInProgress = ref<string | null>(null)
const deleteTarget = ref<AdminBoardInfo | null>(null)
const toastMessage = ref<string>('')
const searchQuery = ref<string>('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const showingStart = computed(() => {
  if (totalBoards.value === 0) return 0
  return (currentPage.value - 1) * 10 + 1
})

const showingEnd = computed(() => {
  return Math.min(currentPage.value * 10, totalBoards.value)
})

onMounted(async () => {
  if (passkey.value) {
    const valid = await verifyAdminPasskey(passkey.value)
    if (valid) {
      isAuthenticated.value = true
      fetchBoards(1)
    } else {
      sessionStorage.removeItem(passkeyKey)
      passkey.value = ''
    }
  }
})

const handleLogin = async () => {
  loginError.value = ''
  isAuthenticating.value = true
  try {
    const valid = await verifyAdminPasskey(inputPasskey.value)
    if (valid) {
      passkey.value = inputPasskey.value
      sessionStorage.setItem(passkeyKey, inputPasskey.value)
      isAuthenticated.value = true
      fetchBoards(1)
    } else {
      loginError.value = 'Invalid Admin Passkey. Please try again.'
    }
  } catch (err) {
    loginError.value = 'Authentication failed. Ensure ADMIN_PASSKEY is configured.'
  } finally {
    isAuthenticating.value = false
  }
}

const logout = () => {
  sessionStorage.removeItem(passkeyKey)
  passkey.value = ''
  isAuthenticated.value = false
}

const fetchBoards = async (page: number) => {
  if (!passkey.value) return
  isLoading.value = true
  try {
    const res = await getAdminBoards(passkey.value, page, searchQuery.value)
    boards.value = res.boards
    totalBoards.value = res.total
    currentPage.value = res.page
    totalPages.value = res.totalPages
  } catch (err) {
    console.error('Failed to fetch admin boards', err)
    if ((err as Error).message.includes('401') || (err as Error).message.includes('Unauthorized')) {
      logout()
    }
  } finally {
    isLoading.value = false
  }
}

const onSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchBoards(1)
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  currentPage.value = 1
  fetchBoards(1)
}

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchBoards(page)
}

const handleExtendExpiry = async (boardId: string) => {
  actionInProgress.value = boardId
  try {
    const updated = await extendBoardExpiry(passkey.value, boardId)
    showToast(`Expiry extended for board '${updated.name || updated.id}'!`)
    fetchBoards(currentPage.value)
  } catch (err) {
    alert('Failed to extend board expiry: ' + (err as Error).message)
  } finally {
    actionInProgress.value = null
  }
}

const handleRemoveExpiry = async (boardId: string) => {
  actionInProgress.value = boardId
  try {
    const updated = await removeBoardExpiry(passkey.value, boardId)
    showToast(`Expiry removed for board '${updated.name || updated.id}'. Board is now persistent!`)
    fetchBoards(currentPage.value)
  } catch (err) {
    alert('Failed to remove board expiry: ' + (err as Error).message)
  } finally {
    actionInProgress.value = null
  }
}

const confirmDelete = (board: AdminBoardInfo) => {
  deleteTarget.value = board
}

const executeDelete = async () => {
  if (!deleteTarget.value) return
  const boardId = deleteTarget.value.id
  const name = deleteTarget.value.name
  actionInProgress.value = boardId
  try {
    await deleteBoardByAdmin(passkey.value, boardId)
    showToast(`Board '${name || boardId}' deleted successfully.`)
    deleteTarget.value = null
    fetchBoards(currentPage.value)
  } catch (err) {
    alert('Failed to delete board: ' + (err as Error).message)
  } finally {
    actionInProgress.value = null
  }
}

const showToast = (msg: string) => {
  toastMessage.value = msg
  setTimeout(() => {
    if (toastMessage.value === msg) {
      toastMessage.value = ''
    }
  }, 4000)
}

const formatDate = (timestampUtcSec: number) => {
  if (!timestampUtcSec || timestampUtcSec <= 0) return '-'
  const d = new Date(timestampUtcSec * 1000)
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatTTL = (ttlSeconds: number) => {
  if (ttlSeconds <= 0) return 'Expired or Persistent'
  const hours = Math.floor(ttlSeconds / 3600)
  const minutes = Math.floor((ttlSeconds % 3600) / 60)
  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remHours = hours % 24
    return `Expires in ${days}d ${remHours}h`
  }
  if (hours > 0) {
    return `Expires in ${hours}h ${minutes}m`
  }
  return `Expires in ${minutes}m`
}
</script>
