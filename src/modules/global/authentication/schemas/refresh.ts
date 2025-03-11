import { createZodDto } from '@anatine/zod-nestjs';
import { z } from 'zod';

export const refreshTokenSchema = z.object({
  refreshToken: z.string({
    required_error: 'Token é obrigatório',
  }),
});

export type RefreshTokenSchemaData = z.infer<typeof refreshTokenSchema>;
export class RefreshTokenDto extends createZodDto(refreshTokenSchema) {}
