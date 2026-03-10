<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const { signIn, signUp, isAuthenticated } = useAuth()
const email = ref('')
const password = ref('')
const isSignUp = ref(false)
const error = ref('')
const loading = ref(false)

onMounted(() => {
  if (isAuthenticated.value) navigateTo('/dashboard')
})

const submit = async () => {
  error.value = ''
  loading.value = true
  try {
    if (isSignUp.value) {
      await signUp(email.value, password.value)
    } else {
      await signIn(email.value, password.value)
    }
    navigateTo('/dashboard')
  } catch (e: any) {
    error.value = e?.data?.error || 'Authentication failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 text-center mb-2">Phone Call Receptionist</h1>
    <p class="text-sm text-gray-500 text-center mb-8">AI Voice Assistant for IT Services</p>

    <form @submit.prevent="submit" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input v-model="email" type="email" required
          class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input v-model="password" type="password" required minlength="8"
          class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent" />
      </div>

      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

      <button type="submit" :disabled="loading"
        class="w-full py-2 px-4 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 font-medium">
        {{ loading ? 'Loading...' : (isSignUp ? 'Sign Up' : 'Sign In') }}
      </button>
    </form>

    <p class="text-center text-sm text-gray-500 mt-4">
      {{ isSignUp ? 'Already have an account?' : "Don't have an account?" }}
      <button @click="isSignUp = !isSignUp" class="text-primary-600 hover:underline ml-1">
        {{ isSignUp ? 'Sign In' : 'Sign Up' }}
      </button>
    </p>
  </div>
</template>
