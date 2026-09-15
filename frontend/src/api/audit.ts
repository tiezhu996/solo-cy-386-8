import request from '../utils/request'
import type { AuditVO, PageResult } from './types'

export function listAudits(params: Record<string, unknown> = {}) {
  return request.get('/audits', { params })
}

export type { AuditVO, PageResult }
