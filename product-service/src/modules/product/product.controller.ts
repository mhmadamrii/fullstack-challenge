import { Body, Controller, Get, Post, HttpCode, Param } from '@nestjs/common';
import { ProductService } from './product.service';
import { CreateProductDto } from './product.dto';

@Controller('products')
export class ProductController {
  constructor(private readonly productService: ProductService) {}

  @Get()
  async getProducts() {
    return this.productService.getAllProducts();
  }

  @Get('no-cache')
  async getProductsWithoutCache() {
    return this.productService.getAllProductsWithoutCache();
  }

  @Get(':id')
  async getProductById(@Param('id') id: string) {
    return this.productService.getProductById(id);
  }

  @Post()
  @HttpCode(201)
  async createProduct(@Body() createProductDto: CreateProductDto) {
    return this.productService.createProduct(createProductDto);
  }

  @Post('no-cache')
  @HttpCode(201)
  async createProductWithoutCache(@Body() createProductDto: CreateProductDto) {
    return this.productService.createProductWithoutCache(createProductDto);
  }
}
