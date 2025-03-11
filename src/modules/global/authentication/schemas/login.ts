import { createZodDto } from '@anatine/zod-nestjs';
import { z } from 'zod';

export const loginSchema = z.object({
  email: z
    .string({
      required_error: 'Email é obrigatório',
    })
    .email('Email inválido'),
  password: z.string({
    required_error: 'Senha é obrigatória',
  }),
});

export type LoginSchemaData = z.infer<typeof loginSchema>;
export class LoginDto extends createZodDto(loginSchema) {}
