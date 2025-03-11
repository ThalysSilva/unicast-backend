import { registerAs } from '@nestjs/config';
import { jwtConstants } from '../constants';
import { JwtSignOptions } from '@nestjs/jwt';

export default registerAs(
  'refresh-jwt',
  (): JwtSignOptions => ({
    secret: jwtConstants.refreshTokenSecretKey,
    expiresIn: jwtConstants.refreshTokenExpiresIn,
  }),
);
