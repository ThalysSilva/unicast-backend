import { Module } from '@nestjs/common';
import { LocalStrategy } from './strategies/local.strategy';
import { PassportModule } from '@nestjs/passport';
import { JwtModule } from '@nestjs/jwt';
import { JwtStrategy } from './strategies/jwt.strategy';
import { AuthenticationController } from './authentication.controller';
import jwtConfig from './config/jwt.config';
import refreshJwtConfig from './config/refresh-jwt.config';
import { ConfigModule } from '@nestjs/config';
import { AuthenticationService } from './authentication.service';
import { RefreshJwtStrategy } from './strategies/refresh-jwt.strategy';
import { UserModule } from 'src/modules/user/user.module';

@Module({
  controllers: [AuthenticationController],
  imports: [
    PassportModule,
    UserModule,
    JwtModule.registerAsync(jwtConfig.asProvider()),
    ConfigModule.forRoot({
      isGlobal: true,
      load: [refreshJwtConfig, jwtConfig],
    }),
  ],
  providers: [
    AuthenticationService,
    LocalStrategy,
    JwtStrategy,
    RefreshJwtStrategy,
  ],
  exports: [AuthenticationService],
})
export class AuthenticationModule {}
