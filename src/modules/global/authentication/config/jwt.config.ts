import { registerAs } from '@nestjs/config';
import { jwtConstants } from '../constants';
import { JwtModuleOptions } from '@nestjs/jwt';

export default registerAs(
  'jwt',
  (): JwtModuleOptions => ({
    secret: jwtConstants.tokenSecretKey,
    signOptions: { expiresIn: jwtConstants.tokenExpiresIn },
  }),
);
