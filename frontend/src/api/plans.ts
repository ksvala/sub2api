import { apiClient } from './client'
import type { PaginatedResponse, Plan } from '@/types'

export const plansAPI = {
  async getPlans() {
    const { data } = await apiClient.get<PaginatedResponse<Plan>>('/plans', {
      params: { page: 1, page_size: 200 }
    })
    return data.items || []
  }
}
