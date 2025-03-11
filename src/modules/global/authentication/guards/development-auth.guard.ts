import {
  Injectable,
  CanActivate,
  ExecutionContext,
  UnauthorizedException,
} from '@nestjs/common';
import { tokens } from '../constants';

@Injectable()
export class DevelopmentAuthGuard implements CanActivate {
  private readonly validToken = tokens.development;
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest<Request>();

    const authorizationHeader = request.headers['authorization'];

    if (!authorizationHeader) {
      throw new UnauthorizedException('Autenticação não encontrada');
    }

    const token = authorizationHeader.split(' ')[1];
    if (token !== this.validToken) {
      throw new UnauthorizedException('Autenticação inválida');
    }

    return true;
  }
}
